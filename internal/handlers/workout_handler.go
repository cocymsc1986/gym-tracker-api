package handlers

import (
	"gym-tracker-api/internal/models"
	"gym-tracker-api/internal/services"
	"gym-tracker-api/internal/utils"
	"net/http"

	"github.com/gorilla/mux"
)

type WorkoutHandler struct {
	service services.WorkoutService
}

func NewWorkoutHandler(service services.WorkoutService) *WorkoutHandler {
	return &WorkoutHandler{
		service: service,
	}
}

func (h *WorkoutHandler) GetWorkout(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]
	workoutID := vars["workoutId"]

	workout, err := h.service.GetWorkout(userID, workoutID)
	if err != nil {
		utils.WriteErrorResponse(w, err)
		return
	}
	utils.WriteJSONResponse(w, workout, http.StatusOK)
}

func (h *WorkoutHandler) ListWorkouts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	workouts, err := h.service.GetWorkouts(userID)
	if err != nil {
		utils.WriteErrorResponse(w, err)
		return
	}

	utils.WriteJSONResponse(w, workouts, http.StatusOK)
}

func (h *WorkoutHandler) CreateWorkout(w http.ResponseWriter, r *http.Request) {
	var workout models.Workout
	workout.WorkoutID = utils.GenerateUUID()
	workout.CreatedAt = utils.GetCurrentTime()
	workout.UserID = mux.Vars(r)["userId"]
	workout.Exercises = []string{}

	if err := utils.DecodeJSON(r.Body, &workout); err != nil {
		utils.WriteErrorResponse(w, err)
		return
	}

	if err := h.service.CreateWorkout(&workout); err != nil {
		utils.WriteErrorResponse(w, err)
		return
	}

	utils.WriteJSONResponse(w, workout, http.StatusCreated)
}

func (h *WorkoutHandler) UpdateWorkout(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]
	workoutID := vars["workoutId"]

	var workout models.Workout
	if err := utils.DecodeJSON(r.Body, &workout); err != nil {
		utils.WriteErrorResponse(w, err)
		return
	}

	if err := h.service.UpdateWorkout(userID, workoutID, &workout); err != nil {
		utils.WriteErrorResponse(w, err)
		return
	}

	utils.WriteJSONResponse(w, workout, http.StatusOK)
}

func (h *WorkoutHandler) DeleteWorkout(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]
	workoutID := vars["workoutId"]

	if err := h.service.DeleteWorkout(userID, workoutID); err != nil {
		utils.WriteErrorResponse(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *WorkoutHandler) AddExerciseToWorkout(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]
	workoutID := vars["workoutId"]
	exerciseID := vars["exerciseId"]

	if err := h.service.AddExerciseToWorkout(userID, workoutID, exerciseID); err != nil {
		utils.WriteErrorResponse(w, err)
		return
	}

	utils.WriteJSONResponse(w, exerciseID, http.StatusCreated)
}

func (h *WorkoutHandler) RemoveExerciseFromWorkout(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]
	workoutID := vars["workoutId"]
	exerciseID := vars["exerciseId"]

	if err := h.service.RemoveExerciseFromWorkout(userID, workoutID, exerciseID); err != nil {
		utils.WriteErrorResponse(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *WorkoutHandler) ListExercisesInWorkout(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]
	workoutID := vars["workoutId"]

	workout, err := h.service.GetWorkout(workoutID, userID)
	if err != nil {
		utils.WriteErrorResponse(w, err)
		return
	}
	exercises := workout.Exercises
	if exercises == nil {
		exercises = []string{}
	}

	utils.WriteJSONResponse(w, exercises, http.StatusOK)
}
