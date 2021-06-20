package main

import (
	"errors"
	"finalTask/internal/data"
	"finalTask/internal/validator"
	"fmt"
	"net/http"
)

func(app*application) createMangaHandler(w http.ResponseWriter,r *http.Request){
	var input struct{
		Title string `json:"title"`
		Year  int32 `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres []string `json:"genres"`
	}

	err:=app.readJSON(w,r,&input)
	if err!=nil{
		app.badRequestResponse(w,r,err)
		return
	}

	manga:=&data.Manga{
		Title:input.Title,
		Year:input.Year,
		Runtime:input.Runtime,
		Genres:input.Genres,
	}

	v:=validator.New()

	if data.ValidateManga(v,manga); !v.Valid(){
		app.failedValidationResponse(w,r,v.Errors)
		return
	}

	err=app.models.Manga.Insert(manga)
	if err!=nil{
		app.serverErrorResponse(w,r,err)
		return
	}

	headers:=make(http.Header)
	headers.Set("Location",fmt.Sprintf("/v1/manga/%d",manga.ID))

	err=app.writeJSON(w,http.StatusCreated,envelope{"manga":manga},headers)
	if err!=nil{
		app.serverErrorResponse(w,r,err)
	}
}

func(app*application) showMangaHandler(w http.ResponseWriter, r*http.Request){
	id,err:=app.readIDParam(r)
	if err!=nil{
		http.NotFound(w,r)
		return
	}

	manga,err:=app.models.Manga.Get(id)
	if err!=nil{
		switch{
		case errors.Is(err,data.ErrRecordNotFound):
			app.notFoundResponse(w,r)
		default:
			app.serverErrorResponse(w,r,err)
		}
		return
	}

	err=app.writeJSON(w,http.StatusOK,envelope{"manga":manga},nil)
	if err!=nil{
		app.serverErrorResponse(w,r,err)
	}
}

func(app *application)updateMovieHandler(w http.ResponseWriter,r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	manga, err := app.models.Manga.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}


	var input struct{
		Title *string `json:"title"`
		Year *int32 `json:"year"`
		Runtime *data.Runtime `json:"runtime"`
		Genres []string `json:"genres"`
	}

	err=app.readJSON(w,r,&input)
	if err!=nil{
		app.badRequestResponse(w,r,err)
		return
	}

	if input.Year!=nil{
		manga.Year=*input.Year
	}

	if input.Runtime!=nil{
		manga.Runtime=*input.Runtime
	}

	if input.Genres!=nil{
		manga.Genres=input.Genres
	}

	v:=validator.New()

	if data.ValidateManga(v,manga); !v.Valid(){
		app.failedValidationResponse(w,r,v.Errors)
		return
	}

	err=app.models.Manga.Update(manga)
	if err!=nil{
		app.serverErrorResponse(w,r,err)
		return
	}

	err=app.writeJSON(w,http.StatusOK,envelope{"manga":manga},nil)
	if err!=nil{
		app.serverErrorResponse(w,r,err)
	}
}

func(app *application)deleteMovieHandler(w http.ResponseWriter,r *http.Request){
	id,err:=app.readIDParam(r)
	if err!=nil{
		app.notFoundResponse(w,r)
		return
	}

	err=app.models.Manga.Delete(id)
	if err!=nil{
		switch{
		case errors.Is(err,data.ErrRecordNotFound):
			app.notFoundResponse(w,r)
		default:
			app.serverErrorResponse(w,r,err)
		}
		return
	}

	err=app.writeJSON(w,http.StatusOK,envelope{"message":"manga successfully deleted"},nil)
	if err!=nil{
		app.serverErrorResponse(w,r,err)
	}
}

func(app *application)listMoviesHandler(w http.ResponseWriter,r *http.Request){
	var input struct{
		Title string
		Genres []string
		data.Filters
	}

	v:=validator.New()

	qs:=r.URL.Query()

	input.Title=app.readString(qs,"title","")
	input.Genres=app.readCSV(qs,"genres",[]string{})

	input.Filters.Page=app.readInt(qs,"page",1,v)
	input.Filters.PageSize=app.readInt(qs,"page_size",20,v)

	input.Filters.Sort=app.readString(qs,"sort","id")
	input.Filters.SortSafelist=[]string{"id","title","year","runtime","-id","-title","-year","-runtime"}

	if data.ValidateFilters(v,input.Filters); !v.Valid(){
		app.failedValidationResponse(w,r,v.Errors)
		return
	}

	manga,metadata,err:=app.models.Manga.GetAll(input.Title,input.Genres,input.Filters)
	if err!=nil{
		app.serverErrorResponse(w,r,err)
		return
	}

	err=app.writeJSON(w,http.StatusOK,envelope{"manga":manga,"metadata": metadata},nil)
	if err!=nil{app.serverErrorResponse(w,r,err)}
	fmt.Fprintf(w,"%+v\n",input)
}