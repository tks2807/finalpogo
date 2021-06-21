package data

import (
	"context"
	"database/sql"
	"errors"
	"finalTask/internal/validator"
	"fmt"
	"github.com/lib/pq"
	"time"
)

type Manga struct {
	ID int64 `json:"id"`
	CreatedAt time.Time `json:"-"`
	Title string `json:"title"`
	Year int32 `json:"year,omitempty"`
	Runtime Runtime `json:"runtime,omitempty"`
	Genres []string `json:"genres,omitempty"`
	Version int32 `json:"version"`
}

func ValidateManga(v *validator.Validator, manga *Manga) {
	v.Check(manga.Title != "", "title", "must be provided")
	v.Check(len(manga.Title) <= 500, "title", "must not be more than 500 bytes long")

	v.Check(manga.Year != 0, "year", "must be provided")
	v.Check(manga.Year >= 1888, "year", "must be greater than 1888")
	v.Check(manga.Year <= int32(time.Now().Year()), "year", "must not be in the future")
}

type MangaModel struct{
	DB *sql.DB
}

func(m MangaModel)Insert(manga *Manga)error{
	query:= `
		INSERT INTO manga (title, year, runtime, genres)         
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, version`

	args:=[]interface{}{manga.Title,manga.Year,manga.Runtime,pq.Array(manga.Genres)}

	ctx,cancel:=context.WithTimeout(context.Background(),3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx,query,args...).Scan(&manga.ID,&manga.CreatedAt,&manga.Version)
}

func(m MangaModel)Get(id int64)(*Manga, error){
	if id<1{
		return nil,ErrRecordNotFound
	}
	query:=`
		SELECT id, created_at, title, year, runtime, genres, version
		FROM manga        
		WHERE id = $1`

	var manga Manga

	ctx,cancel:=context.WithTimeout(context.Background(),3*time.Second)

	defer cancel()

	err:=m.DB.QueryRowContext(ctx,query,id).Scan(
		&manga.ID,
		&manga.CreatedAt,
		&manga.Title,
		&manga.Year,
		&manga.Runtime,pq.Array(&manga.Genres),
		&manga.Version,
		)

	if err!=nil{
		switch{
		case errors.Is(err,sql.ErrNoRows):
			return nil,ErrRecordNotFound

		default:
			return nil,err
		}
	}
	return &manga,nil
}

func(m MangaModel)Update(manga *Manga)error{
	query:=`
		UPDATE manga
		SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
		WHERE id = $5 AND version = $6
		RETURNING version`

	args:=[]interface{}{
		manga.Title,
		manga.Year,
		manga.Runtime,pq.Array(manga.Genres),
		manga.ID,
		manga.Version,
	}

	ctx,cancel:=context.WithTimeout(context.Background(),3*time.Second)
	defer cancel()

	err:=m.DB.QueryRowContext(ctx,query,args...).Scan(&manga.Version)
	if err!=nil{
		switch{
		case errors.Is(err,sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

func(m MangaModel)Delete(id int64)error{
	if id<1{
		return ErrRecordNotFound
	}

	query:=`       
		DELETE FROM manga
		WHERE id = $1`

	ctx,cancel:=context.WithTimeout(context.Background(),3*time.Second)
	defer cancel()

	result,err:=m.DB.ExecContext(ctx,query,id)
	if err!=nil{
		return err
	}

	rowsAffected,err:=result.RowsAffected()
	if err!=nil{
		return err
	}

	if rowsAffected==0{
		return ErrRecordNotFound
	}

	return nil
}

func(m MangaModel)GetAll(title string,genres []string,filters Filters)([]*Manga,Metadata,error){
	query:=fmt.Sprintf(`       
		SELECT count(*) OVER(), id, created_at, title, year, runtime, genres, version
		FROM manga   
		WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '') 
		AND (genres @> $2 OR $2 = '{}')
		ORDER BY %s %s, id ASC
		LIMIT $3 OFFSET $4`,filters.sortColumn(),filters.sortDirection())

	ctx,cancel:=context.WithTimeout(context.Background(),3*time.Second)
	defer cancel()

	args:=[]interface{}{title,pq.Array(genres),filters.limit(),filters.offset()}

	rows,err:=m.DB.QueryContext(ctx,query,args...)
	if err!=nil{
		return nil,Metadata{},err
	}

	defer rows.Close()
	totalRecords:=0
	mangas:=[]*Manga{}
	for rows.Next(){
		var manga Manga

		err:=rows.Scan(
			&totalRecords,
			&manga.ID,
			&manga.CreatedAt,
			&manga.Title,
			&manga.Year,
			&manga.Runtime,pq.Array(&manga.Genres),
			&manga.Version,
			)
		if err!=nil{
			return nil,Metadata{},err
		}
		mangas=append(mangas,&manga)
	}
	if err=rows.Err(); err!=nil{
		return nil,Metadata{},err
	}
	metadata:=calculateMetadata(totalRecords,filters.Page,filters.PageSize)
	return mangas,metadata,nil
}
