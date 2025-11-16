package handlers

import (
	"gym-tracker-api/internal/models"
	"gym-tracker-api/internal/services"
	"gym-tracker-api/internal/utils"
	"net/http"

	"github.com/gorilla/mux"
)

type ExerciseHandler struct {
	service services.ExerciseService
}

func NewExerciseHandler(service services.ExerciseService) *ExerciseHandler {
	return &ExerciseHandler{
		service: service,
	}
}

func (h *ExerciseHandler) GetExercises(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userId"]

	exercises, err := h.service.GetExercises(userID)
	if err != nil {
		utils.WriteErrorResponse(w, err)
		return
	}
	utils.WriteJSONResponse(w, exercises, http.StatusOK)
}

func (h *ExerciseHandler) GetExercise(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]
	exerciseID := vars["exerciseId"]

	exercise, err := h.service.GetExercise(userID, exerciseID)
	if err != nil {
		utils.WriteErrorResponse(w, err)
		return
	}
	utils.WriteJSONResponse(w, exercise, http.StatusOK)
}

func (h *ExerciseHandler) ListExercisesByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]
	exerciseName := vars["exerciseName"]

	exercise, err := h.service.ListExercisesByName(userID, exerciseName)
	if err != nil {
		utils.WriteErrorResponse(w, err)
		return
	}
	utils.WriteJSONResponse(w, exercise, http.StatusOK)
}

func (h *ExerciseHandler) CreateExercise(w http.ResponseWriter, r *http.Request) {
	var exercise models.Exercise
	vars := mux.Vars(r)
	userId := vars["userId"]
	if err := utils.DecodeJSON(r.Body, &exercise); err != nil {
		utils.WriteErrorResponse(w, err)
		return
	}

	if err := h.service.CreateExercise(userId, &exercise); err != nil {
		utils.WriteErrorResponse(w, err)
		return
	}

	utils.WriteJSONResponse(w, exercise, http.StatusCreated)
}

func (h *ExerciseHandler) UpdateExercise(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]
	exerciseID := vars["exerciseId"]

	var exercise models.Exercise
	if err := utils.DecodeJSON(r.Body, &exercise); err != nil {
		utils.WriteErrorResponse(w, err)
		return
	}

	if err := h.service.UpdateExercise(userID, exerciseID, &exercise); err != nil {
		utils.WriteErrorResponse(w, err)
		return
	}

	utils.WriteJSONResponse(w, exercise, http.StatusOK)
}

func (h *ExerciseHandler) DeleteExercise(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]
	exerciseID := vars["exerciseId"]

	if err := h.service.DeleteExercise(userID, exerciseID); err != nil {
		utils.WriteErrorResponse(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
