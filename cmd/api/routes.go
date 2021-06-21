package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func(app*application) routes() http.Handler{
	router:=httprouter.New()

	router.NotFound=http.HandlerFunc(app.notFoundResponse)

	router.MethodNotAllowed=http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet,"/v1/mangacheck",app.mangacheckHandler)
	router.HandlerFunc(http.MethodGet,"/v1/mangas",app.listMoviesHandler)
	router.HandlerFunc(http.MethodPost,"/v1/manga",app.createMangaHandler)
	router.HandlerFunc(http.MethodGet,"/v1/manga/:id",app.showMangaHandler)
	router.HandlerFunc(http.MethodPatch,"/v1/manga/:id",app.updateMovieHandler)
	router.HandlerFunc(http.MethodDelete,"/v1/manga/:id",app.deleteMovieHandler)
	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)

	return app.recoverPanic(app.rateLimit(router))
}

