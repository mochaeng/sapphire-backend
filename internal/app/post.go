package app

import (
	"errors"
	"net/http"

	"github.com/mochaeng/sapphire-backend/internal/config"
	"github.com/mochaeng/sapphire-backend/internal/httpio"
	"github.com/mochaeng/sapphire-backend/internal/models"
	service "github.com/mochaeng/sapphire-backend/internal/services"
	"github.com/mochaeng/sapphire-backend/internal/store"
)

func getPostFromCtx(r *http.Request) *models.Post {
	post, _ := r.Context().Value(postCtx).(*models.Post)
	return post
}

// CreatePost godoc
//
//	@Summary		Creates a post
//	@Description	A activated and authenticated user can create a post
//	@Tags			post
//	@Accept			mpfd
//	@Produce		json
//	@Param			tittle	formData	string	true	"Post tittle"
//	@Param			content	formData	string	true	"Post content"
//	@Param			media	formData	file	false	"Post media"
//	@Success		201		{object}	models.CreatePostResponse
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/post [post]
func (app *Application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(config.MaxMediaUploadSize)
	if err != nil {
		app.BadRequestResponse(w, r, err)
		return
	}

	var payload models.CreatePostPayload
	if err := httpio.ReadFormDataValues(r, &payload); err != nil {
		app.BadRequestResponse(w, r, err)
		return
	}

	fileField := "media"
	file, err := httpio.ReadFormFile(r, fileField, config.MaxMediaUploadSize)
	if err != nil {
		app.BadRequestResponse(w, r, err)
		return
	}

	user := getUserFromContext(r)
	post, err := app.Service.Post.Create(r.Context(), user, &payload, file)
	if err != nil {
		switch err {
		case service.ErrInvalidPayload, service.ErrSaveFile:
			app.BadRequestResponse(w, r, err)
		case store.ErrNotFound:
			app.NotFoundResponse(w, r, err)
		default:
			app.InternalServerErrorResponse(w, r, err)
		}
		return
	}

	response := &models.CreatePostResponse{
		ID:        post.ID,
		Tittle:    post.Tittle,
		Content:   post.Content,
		Tags:      post.Tags,
		MediaURL:  post.Media,
		CreatedAt: post.CreatedAt,
		UserID:    user.ID,
	}
	if err := httpio.JsonResponse(w, http.StatusCreated, response); err != nil {
		app.InternalServerErrorResponse(w, r, err)
		return
	}
}

// GetPost godoc
//
//	@Summary		Gets a post
//	@Description	Gets a post by its own ID
//	@Tags			post
//	@Accept			json
//	@Produce		json
//	@Param			postID	path		string	true	"Post ID"
//	@Success		200		{object}	models.GetPostResponse
//	@Failure		400		{object}	error
//	@Failure		404		{object}	error
//	@Failure		500		{object}	error
//	@Router			/post/{postID} [get]
func (app *Application) getPostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)
	response := &models.GetPostResponse{
		Tittle:    post.Tittle,
		Content:   post.Content,
		Tags:      post.Tags,
		MediaURL:  post.Media,
		CreatedAt: post.CreatedAt,
		UpdatedAt: post.CreatedAt,
		User: models.UserResponse{
			ID:        post.User.ID,
			Username:  post.User.Username,
			FirstName: post.User.FirstName,
			LastName:  post.User.LastName,
		},
	}
	if err := httpio.JsonResponse(w, http.StatusOK, response); err != nil {
		app.InternalServerErrorResponse(w, r, err)
		return
	}
}

// DeletePost godoc
//
//	@Summary		Deletes a post
//	@Description	Delete a post by ID
//	@Tags			post
//	@Accept			json
//	@Produce		json
//	@Param			postID	path		int	true	"Post ID"
//	@Success		204		{object}	string
//	@Failure		404		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/post/{postID} [delete]
func (app *Application) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)
	if err := app.Service.Post.Delete(r.Context(), post.ID); err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.NotFoundResponse(w, r, err)
		default:
			app.InternalServerErrorResponse(w, r, err)
		}
		return
	}
	httpio.NoContentResponse(w)
}

// UpdatePost godoc
//
//	@Summary		Updates a post
//	@Description	Allows a user to update their own post
//	@Tags			post
//	@Accept			json
//	@Produce		json
//	@Param			postID	path		string						true	"Post ID"
//	@Param			payload	body		models.UpdatePostPayload	true	"Update post payload"
//	@Success		200		{object}	models.UpdatePostResponse
//	@Failure		400		{object}	error
//	@Failure		404		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/post/{postID} [patch]
func (app *Application) updatePostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)
	var payload models.UpdatePostPayload
	if err := httpio.ReadJSON(w, r, &payload); err != nil {
		app.BadRequestResponse(w, r, err)
		return
	}

	if err := app.Service.Post.Update(r.Context(), post, &payload); err != nil {
		switch err {
		case store.ErrNotFound:
			app.NotFoundResponse(w, r, err)
		default:
			app.InternalServerErrorResponse(w, r, err)
		}
		return
	}

	response := &models.UpdatePostResponse{
		Tittle:    post.Tittle,
		Content:   post.Content,
		UpdatedAt: post.UpdatedAt,
	}
	if err := httpio.JsonResponse(w, http.StatusOK, response); err != nil {
		app.InternalServerErrorResponse(w, r, err)
		return
	}
}
