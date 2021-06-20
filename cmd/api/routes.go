package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func(app*application) routes() *httprouter.Router{
	router:=httprouter.New()

	router.NotFound=http.HandlerFunc(app.notFoundResponse)

	router.MethodNotAllowed=http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet,"/v1/mangacheck",app.mangacheckHandler)
	router.HandlerFunc(http.MethodGet,"/v1/movies",app.listMoviesHandler)
	router.HandlerFunc(http.MethodPost,"/v1/manga",app.createMangaHandler)
	router.HandlerFunc(http.MethodGet,"/v1/manga/:id",app.showMangaHandler)
	router.HandlerFunc(http.MethodPatch,"/v1/movies/:id",app.updateMovieHandler)
	router.HandlerFunc(http.MethodDelete,"/v1/movies/:id",app.deleteMovieHandler)

	return router
}

