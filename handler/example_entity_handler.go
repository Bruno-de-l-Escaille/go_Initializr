package handler

import (
	"context"
	"go_Initializr/handler/dto"
	"go_Initializr/models"
	"go_Initializr/service"
	"go_Initializr/service/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// createContextWithUser creates a context with user UUID from Gin context
func createContextWithUser(c *gin.Context) context.Context {
	if userUUID, exists := c.Get("user_uuid"); exists {
		return utils.CreateContextWithUserUUID(c.Request.Context(), userUUID.(string))
	}
	return c.Request.Context()
}

// ResponseWithCount is a custom response structure to include count and data
type ResponseWithCount struct {
	Count int                     `json:"count"`
	Data  []*models.ExampleEntity `json:"data"`
}

// ResponseError represents an error response
type ResponseError struct {
	Error string `json:"error"`
}

// ExampleEntityHandler handles HTTP requests for ExampleEntity
type ExampleEntityHandler struct {
	ExampleEntityService service.ExampleEntityServiceInterface
}

// CreateEntity handles the creation of a new ExampleEntity
// @Summary Create a new ExampleEntity
// @Description Creates a new ExampleEntity with the provided data. Requires JWT authentication.
// @Tags example-entities
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Authorization header string true "Bearer token"
// @Param entity body dto.ExampleEntityCreateRequest true "ExampleEntity data"
// @Success 201 {object} models.ExampleEntity
// @Failure 400 {object} ResponseError "Bad Request - Invalid input data"
// @Failure 401 {object} ResponseError "Unauthorized - Valid JWT token required"
// @Failure 409 {object} ResponseError "Conflict - Entity already exists"
// @Failure 500 {object} ResponseError "Internal Server Error"
// @Router /api/example-entities [post]
func (h *ExampleEntityHandler) CreateEntity(c *gin.Context) {
	// Get user UUID from JWT context
	userUUID, exists := c.Get("user_uuid")
	if !exists {
		c.JSON(http.StatusUnauthorized, ResponseError{Error: "User not authenticated"})
		return
	}

	var req dto.ExampleEntityCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Str("user_uuid", userUUID.(string)).Msg("Failed to bind JSON in CreateEntity")
		c.JSON(http.StatusBadRequest, ResponseError{Error: "Invalid input data: " + err.Error()})
		return
	}

	// Create a context with the user UUID for the service layer
	ctx := createContextWithUser(c)
	
	entity, err := h.ExampleEntityService.CreateEntity(ctx, models.ExampleEntity{
		Name:        req.Name,
		Description: req.Description,
		Email:       req.Email,
		Status:      req.Status,
	})
	if err != nil {
		log.Error().Err(err).Str("user_uuid", userUUID.(string)).Msg("Failed to create ExampleEntity")
		if err.Error() == "entity with email already exists" {
			c.JSON(http.StatusConflict, ResponseError{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ResponseError{Error: "Failed to create entity"})
		return
	}

	log.Info().Str("entity_uuid", entity.UUID).Str("user_uuid", userUUID.(string)).Msg("ExampleEntity created successfully")
	c.JSON(http.StatusCreated, entity)
}

// GetEntity handles retrieving an ExampleEntity by UUID
// @Summary Get an ExampleEntity by UUID
// @Description Retrieves an ExampleEntity by its UUID. Public endpoint, no authentication required.
// @Tags example-entities
// @Produce json
// @Param uuid path string true "ExampleEntity UUID"
// @Success 200 {object} models.ExampleEntity
// @Failure 400 {object} ResponseError "Bad Request - Invalid UUID"
// @Failure 404 {object} ResponseError "Not Found - Entity not found"
// @Failure 500 {object} ResponseError "Internal Server Error"
// @Router /api/public/example-entities/{uuid} [get]
func (h *ExampleEntityHandler) GetEntity(c *gin.Context) {
	uuid := c.Param("uuid")

	if uuid == "" {
		c.JSON(http.StatusBadRequest, ResponseError{Error: "UUID is required"})
		return
	}

	entity, err := h.ExampleEntityService.GetEntityByRef(c.Request.Context(), uuid)
	if err != nil {
		log.Error().Err(err).Str("uuid", uuid).Msg("Failed to get ExampleEntity")
		c.JSON(http.StatusNotFound, ResponseError{Error: "Entity not found"})
		return
	}

	c.JSON(http.StatusOK, entity)
}

// UpdateEntity handles updating an ExampleEntity
// @Summary Update an ExampleEntity
// @Description Updates an existing ExampleEntity with the provided data. Requires JWT authentication.
// @Tags example-entities
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Authorization header string true "Bearer token"
// @Param uuid path string true "ExampleEntity UUID"
// @Param entity body dto.ExampleEntityUpdateRequest true "Updated ExampleEntity data"
// @Success 200 {object} models.ExampleEntity
// @Failure 400 {object} ResponseError "Bad Request - Invalid input data"
// @Failure 401 {object} ResponseError "Unauthorized - Valid JWT token required"
// @Failure 404 {object} ResponseError "Not Found - Entity not found"
// @Failure 409 {object} ResponseError "Conflict - Email already in use"
// @Failure 500 {object} ResponseError "Internal Server Error"
// @Router /api/example-entities/{uuid} [put]
func (h *ExampleEntityHandler) UpdateEntity(c *gin.Context) {
	// Get user UUID from JWT context
	userUUID, exists := c.Get("user_uuid")
	if !exists {
		c.JSON(http.StatusUnauthorized, ResponseError{Error: "User not authenticated"})
		return
	}

	uuid := c.Param("uuid")
	if uuid == "" {
		c.JSON(http.StatusBadRequest, ResponseError{Error: "UUID is required"})
		return
	}

	var req dto.ExampleEntityUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Str("user_uuid", userUUID.(string)).Str("entity_uuid", uuid).Msg("Failed to bind JSON in UpdateEntity")
		c.JSON(http.StatusBadRequest, ResponseError{Error: "Invalid input data: " + err.Error()})
		return
	}

	// Create a context with the user UUID for the service layer
	ctx := createContextWithUser(c)

	entityData := make(map[string]interface{})
	if req.Name != nil {
		entityData["name"] = *req.Name
	}
	if req.Description != nil {
		entityData["description"] = *req.Description
	}
	if req.Email != nil {
		entityData["email"] = *req.Email
	}
	if req.Status != nil {
		entityData["status"] = *req.Status
	}
	entity, err := h.ExampleEntityService.UpdateEntityByRef(ctx, uuid, entityData)
	if err != nil {
		log.Error().Err(err).Str("user_uuid", userUUID.(string)).Str("entity_uuid", uuid).Msg("Failed to update ExampleEntity")
		if err.Error() == "entity not found" {
			c.JSON(http.StatusNotFound, ResponseError{Error: "Entity not found"})
			return
		}
		if err.Error() == "email is already in use" {
			c.JSON(http.StatusConflict, ResponseError{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ResponseError{Error: "Failed to update entity"})
		return
	}

	log.Info().Str("entity_uuid", uuid).Str("user_uuid", userUUID.(string)).Msg("ExampleEntity updated successfully")
	c.JSON(http.StatusOK, entity)
}

// DeleteEntity handles deleting an ExampleEntity
// @Summary Delete an ExampleEntity
// @Description Deletes an ExampleEntity by its UUID. Requires JWT authentication.
// @Tags example-entities
// @Produce json
// @Security BearerAuth
// @Param Authorization header string true "Bearer token"
// @Param uuid path string true "ExampleEntity UUID"
// @Success 204 "No Content - Entity deleted successfully"
// @Failure 400 {object} ResponseError "Bad Request - Invalid UUID"
// @Failure 401 {object} ResponseError "Unauthorized - Valid JWT token required"
// @Failure 404 {object} ResponseError "Not Found - Entity not found"
// @Failure 500 {object} ResponseError "Internal Server Error"
// @Router /api/example-entities/{uuid} [delete]
func (h *ExampleEntityHandler) DeleteEntity(c *gin.Context) {
	// Get user UUID from JWT context
	userUUID, exists := c.Get("user_uuid")
	if !exists {
		c.JSON(http.StatusUnauthorized, ResponseError{Error: "User not authenticated"})
		return
	}

	uuid := c.Param("uuid")
	if uuid == "" {
		c.JSON(http.StatusBadRequest, ResponseError{Error: "UUID is required"})
		return
	}

	// Create a context with the user UUID for the service layer
	ctx := createContextWithUser(c)
	
	err := h.ExampleEntityService.DeleteEntityByRef(ctx, uuid)
	if err != nil {
		log.Error().Err(err).Str("user_uuid", userUUID.(string)).Str("entity_uuid", uuid).Msg("Failed to delete ExampleEntity")
		if err.Error() == "entity not found" {
			c.JSON(http.StatusNotFound, ResponseError{Error: "Entity not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ResponseError{Error: "Failed to delete entity"})
		return
	}

	log.Info().Str("entity_uuid", uuid).Str("user_uuid", userUUID.(string)).Msg("ExampleEntity deleted successfully")
	c.Status(http.StatusNoContent)
}

// GetAllEntities handles retrieving all ExampleEntities with optional filtering and pagination
// @Summary Get all ExampleEntities
// @Description Retrieves a list of all ExampleEntities with optional filtering and pagination. Public endpoint, no authentication required.
// @Tags example-entities
// @Produce json
// @Param offset query int false "Offset for pagination" default(0)
// @Param limit query int false "Limit for pagination (default: 50, max: 100)"
// @Param filter query string false "Filter string for searching entities"
// @Success 200 {object} ResponseWithCount
// @Failure 400 {object} ResponseError "Bad Request - Invalid query parameters"
// @Failure 500 {object} ResponseError "Internal Server Error"
// @Router /api/public/example-entities [get]
func (h *ExampleEntityHandler) GetAllEntities(c *gin.Context) {
	// Parse query parameters
	offsetStr := c.DefaultQuery("offset", "0")
	limitStr := c.DefaultQuery("limit", "50")
	filter := c.Query("filter")

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		c.JSON(http.StatusBadRequest, ResponseError{Error: "Invalid offset parameter"})
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		c.JSON(http.StatusBadRequest, ResponseError{Error: "Invalid limit parameter"})
		return
	}

	entities, _, totalCount, err := h.ExampleEntityService.GetAllEntities(c.Request.Context(), filter, offset, limit)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get all ExampleEntities")
		c.JSON(http.StatusInternalServerError, ResponseError{Error: "Failed to get entities"})
		return
	}

	// Convert []models.ExampleEntity to []*models.ExampleEntity
	entityPointers := make([]*models.ExampleEntity, len(entities))
	for i := range entities {
		entityPointers[i] = &entities[i]
	}

	response := ResponseWithCount{
		Count: int(totalCount),
		Data:  entityPointers,
	}

	c.JSON(http.StatusOK, response)
}
