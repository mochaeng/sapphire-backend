package app

import (
	"errors"
	"net/http"
	"path/filepath"

	"github.com/mochaeng/sapphire-backend/internal/httpio"
	"github.com/mochaeng/sapphire-backend/internal/media"
	"github.com/mochaeng/sapphire-backend/internal/models"
	"github.com/mochaeng/sapphire-backend/internal/store"
)

func getPostFromCtx(r *http.Request) *models.Post {
	post, _ := r.Context().Value(postCtx).(*models.Post)
	return post
}

// CreatePost godoc
//
//	@Summary		Creates a post
//	@Description	Allows a user to create their own post
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
	err := r.ParseMultipartForm(MaxUploadSize)
	if err != nil {
		app.BadRequestResponse(w, r, err)
		return
	}

	var payload models.CreatePostPayload
	if err := httpio.ReadFormDataValue(r, &payload); err != nil {
		switch err {
		case httpio.ErrMarshalData, httpio.ErrWrongParameterType:
			app.InternalServerErrorResponse(w, r, err)
		default:
			app.BadRequestResponse(w, r, err)
		}
		return
	}
	if err := Validate.Struct(payload); err != nil {
		app.BadRequestResponse(w, r, err)
		return
	}

	fileField := "media"
	file, err := httpio.ReadFormFiles(r, fileField, MaxUploadSize)
	if err != nil {
		app.BadRequestResponse(w, r, err)
		return
	}
	fileUrl := ""
	if file != nil {
		filename, err := media.SaveFileToServer(file, app.Config.MediaFolder)
		if err != nil {
			app.InternalServerErrorResponse(w, r, err)
			return
		}
		fileUrl = filepath.Join(app.Config.MediaFolder, filename)
	}

	user := getUserFromContext(r)
	post := &models.Post{
		Tittle:  payload.Tittle,
		Content: payload.Content,
		Media:   fileUrl,
		Tags:    payload.Tags,
		User:    user,
	}
	if err := app.Store.Post.Create(r.Context(), post); err != nil {
		app.InternalServerErrorResponse(w, r, err)
		return
	}

	response := &models.CreatePostResponse{
		Tittle:    post.Tittle,
		Content:   post.Content,
		Tags:      post.Tags,
		MediaURL:  post.Media,
		CreatedAt: post.CreatedAt,
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
	// idParam := chi.URLParam(r, "postID")
	// postID, err := strconv.ParseInt(idParam, 10, 64)
	// if err != nil {
	// 	app.BadRequestResponse(w, r, err)
	// 	return
	// }
	app.Logger.Infow("what", "post", post)
	if err := app.Store.Post.DeleteByID(r.Context(), post.ID); err != nil {
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
	if err := Validate.Struct(payload); err != nil {
		app.BadRequestResponse(w, r, err)
		return
	}
	if payload.Content != "" {
		post.Content = payload.Content
	}
	if payload.Tittle != "" {
		post.Tittle = payload.Tittle
	}
	if err := app.Store.Post.UpdateByID(r.Context(), post); err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
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
