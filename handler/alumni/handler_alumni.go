package handler_alumni

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/fadhln/lms-be/constants"
	"github.com/fadhln/lms-be/delivery/rs"
	"github.com/fadhln/lms-be/model"
	"github.com/fadhln/lms-be/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

func Init(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

type articleRequest struct {
	Title    string `json:"title" binding:"required,max=200"`
	Slug     string `json:"slug" binding:"required,max=220"`
	Category string `json:"category" binding:"required,max=80"`
	Excerpt  string `json:"excerpt"`
	Body     string `json:"body" binding:"required"`
	Status   string `json:"status"`
}

type eventRequest struct {
	Title                string     `json:"title" binding:"required,max=200"`
	Description          string     `json:"description"`
	Category             string     `json:"category" binding:"required,max=80"`
	Venue                string     `json:"venue" binding:"max=200"`
	StartsAt             time.Time  `json:"starts_at" binding:"required"`
	EndsAt               time.Time  `json:"ends_at"`
	RegistrationDeadline *time.Time `json:"registration_deadline"`
	Status               string     `json:"status"`
}

type profileRequest struct {
	BatchYear  int    `json:"batch_year" binding:"required,min=1900,max=2100"`
	City       string `json:"city" binding:"max=120"`
	Occupation string `json:"occupation" binding:"max=160"`
	Bio        string `json:"bio" binding:"max=4000"`
}

type listingRequest struct {
	ListingType string `json:"listing_type" binding:"required,oneof=business career"`
	Name        string `json:"name" binding:"required,max=200"`
	Category    string `json:"category" binding:"required,max=80"`
	Description string `json:"description" binding:"required,max=4000"`
	ContactURL  string `json:"contact_url" binding:"max=500"`
}

type contactRequest struct {
	Name    string `json:"name" binding:"required,max=160"`
	Email   string `json:"email" binding:"required,email,max=254"`
	Message string `json:"message" binding:"required,max=4000"`
}

type publicAlumni struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Avatar     *string   `json:"avatar,omitempty"`
	BatchYear  int       `json:"batch_year"`
	City       string    `json:"city"`
	Occupation string    `json:"occupation"`
	Bio        string    `json:"bio"`
}

type publicListing struct {
	ID          uuid.UUID    `json:"id"`
	ListingType string       `json:"listing_type"`
	Name        string       `json:"name"`
	Category    string       `json:"category"`
	Description string       `json:"description"`
	ContactURL  string       `json:"contact_url"`
	Owner       publicAlumni `json:"owner"`
}

type reviewedProfile struct {
	ID         uuid.UUID `json:"id"`
	AccountID  uuid.UUID `json:"account_id"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	Avatar     *string   `json:"avatar,omitempty"`
	BatchYear  int       `json:"batch_year"`
	City       string    `json:"city"`
	Occupation string    `json:"occupation"`
	Bio        string    `json:"bio"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

func toPublicAlumni(profile model.AlumniProfile) publicAlumni {
	name := ""
	if profile.Account.Name != nil {
		name = *profile.Account.Name
	}
	return publicAlumni{ID: profile.ID, Name: name, Avatar: profile.Account.Avatar, BatchYear: profile.BatchYear, City: profile.City, Occupation: profile.Occupation, Bio: profile.Bio}
}

func toReviewedProfile(profile model.AlumniProfile) reviewedProfile {
	person := toPublicAlumni(profile)
	return reviewedProfile{ID: profile.ID, AccountID: profile.AccountID, Name: person.Name, Email: profile.Account.Email, Avatar: person.Avatar, BatchYear: profile.BatchYear, City: profile.City, Occupation: profile.Occupation, Bio: profile.Bio, Status: profile.Status, CreatedAt: profile.CreatedAt}
}

func respondError(c *gin.Context, err error) {
	rs.ErrorResponse(c, err)
}

func bind[T any](c *gin.Context) (*T, bool) {
	var request T
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, err)
		return nil, false
	}
	return &request, true
}

func pathID(c *gin.Context, key string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(key))
	if err != nil {
		respondError(c, errors.New("invalid ID"))
		return uuid.Nil, false
	}
	return id, true
}

func listQuery(db *gorm.DB, c *gin.Context, searchable ...string) *gorm.DB {
	if category := c.Query("category"); category != "" {
		db = db.Where("category = ?", category)
	}
	if search := strings.TrimSpace(c.Query("q")); search != "" {
		pattern := "%" + search + "%"
		conditions := make([]string, len(searchable))
		values := make([]interface{}, len(searchable))
		for index, field := range searchable {
			conditions[index] = field + " ILIKE ?"
			values[index] = pattern
		}
		db = db.Where("("+strings.Join(conditions, " OR ")+")", values...)
	}
	return db
}

func (h *Handler) ListNews(c *gin.Context) {
	var articles []model.NewsArticle
	db := listQuery(h.db.Model(&model.NewsArticle{}).Where("status = ?", "published"), c, "title", "excerpt", "body")
	if err := db.Order("published_at DESC NULLS LAST").Find(&articles).Error; err != nil {
		respondError(c, err)
		return
	}
	rs.SuccessResponse(c, articles)
}

func (h *Handler) GetNews(c *gin.Context) {
	var article model.NewsArticle
	if err := h.db.Where("slug = ? AND status = ?", c.Param("slug"), "published").First(&article).Error; err != nil {
		respondError(c, err)
		return
	}
	rs.SuccessResponse(c, article)
}

func (h *Handler) ListEvents(c *gin.Context) {
	var events []model.AlumniEvent
	db := listQuery(h.db.Model(&model.AlumniEvent{}).Where("status = ?", "published"), c, "title", "description", "venue")
	if err := db.Order("starts_at ASC").Find(&events).Error; err != nil {
		respondError(c, err)
		return
	}
	rs.SuccessResponse(c, events)
}

func (h *Handler) GetEvent(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	var event model.AlumniEvent
	if err := h.db.Where("id = ? AND status = ?", id, "published").First(&event).Error; err != nil {
		respondError(c, err)
		return
	}
	rs.SuccessResponse(c, event)
}

func (h *Handler) ListAlumni(c *gin.Context) {
	var profiles []model.AlumniProfile
	db := h.db.Model(&model.AlumniProfile{}).Preload("Account").Where("status = ?", "approved")
	if search := strings.TrimSpace(c.Query("q")); search != "" {
		pattern := "%" + search + "%"
		db = db.Joins("JOIN accounts ON accounts.id = alumni_profiles.account_id").Where("(accounts.name ILIKE ? OR alumni_profiles.occupation ILIKE ? OR alumni_profiles.city ILIKE ?)", pattern, pattern, pattern)
	}
	if year := c.Query("batch_year"); year != "" {
		db = db.Where("batch_year = ?", year)
	}
	if err := db.Order("batch_year DESC").Find(&profiles).Error; err != nil {
		respondError(c, err)
		return
	}
	people := make([]publicAlumni, 0, len(profiles))
	for _, profile := range profiles {
		people = append(people, toPublicAlumni(profile))
	}
	rs.SuccessResponse(c, people)
}

func (h *Handler) ListListings(c *gin.Context) {
	var listings []model.BusinessCareerListing
	db := h.db.Model(&model.BusinessCareerListing{}).Preload("AlumniProfile.Account").Where("status = ?", "approved")
	if kind := c.Query("listing_type"); kind != "" {
		db = db.Where("listing_type = ?", kind)
	}
	db = listQuery(db, c, "name", "category", "description")
	if err := db.Order("created_at DESC").Find(&listings).Error; err != nil {
		respondError(c, err)
		return
	}
	items := make([]publicListing, 0, len(listings))
	for _, listing := range listings {
		items = append(items, publicListing{ID: listing.ID, ListingType: listing.ListingType, Name: listing.Name, Category: listing.Category, Description: listing.Description, ContactURL: listing.ContactURL, Owner: toPublicAlumni(listing.AlumniProfile)})
	}
	rs.SuccessResponse(c, items)
}

func (h *Handler) SubmitContact(c *gin.Context) {
	request, ok := bind[contactRequest](c)
	if !ok {
		return
	}
	message := model.ContactMessage{Name: strings.TrimSpace(request.Name), Email: strings.ToLower(strings.TrimSpace(request.Email)), Message: strings.TrimSpace(request.Message), Status: "new"}
	if err := h.db.Create(&message).Error; err != nil {
		respondError(c, err)
		return
	}
	rs.SuccessResponse(c, message, http.StatusCreated)
}

func (h *Handler) MyProfile(c *gin.Context) {
	account, err := util.GetAccountContext(c, constants.ACCOUNT_STUDENT)
	if err != nil {
		respondError(c, err)
		return
	}
	var profile model.AlumniProfile
	if err := h.db.Preload("Account").Where("account_id = ?", account.ID).First(&profile).Error; err != nil {
		respondError(c, err)
		return
	}
	rs.SuccessResponse(c, gin.H{"name": account.Name, "profile": profile})
}

func (h *Handler) SaveMyProfile(c *gin.Context) {
	account, err := util.GetAccountContext(c, constants.ACCOUNT_STUDENT)
	if err != nil {
		respondError(c, err)
		return
	}
	request, ok := bind[profileRequest](c)
	if !ok {
		return
	}
	profile := model.AlumniProfile{
		AccountID:  account.ID,
		BatchYear:  request.BatchYear,
		City:       strings.TrimSpace(request.City),
		Occupation: strings.TrimSpace(request.Occupation),
		Bio:        strings.TrimSpace(request.Bio),
		Status:     "pending",
	}
	if err := h.db.Where("account_id = ?", account.ID).Assign(profile).FirstOrCreate(&profile).Error; err != nil {
		respondError(c, err)
		return
	}
	rs.SuccessResponse(c, profile)
}

func (h *Handler) SubmitListing(c *gin.Context) {
	account, err := util.GetAccountContext(c, constants.ACCOUNT_STUDENT)
	if err != nil {
		respondError(c, err)
		return
	}
	request, ok := bind[listingRequest](c)
	if !ok {
		return
	}
	var profile model.AlumniProfile
	if err := h.db.Where("account_id = ? AND status = ?", account.ID, "approved").First(&profile).Error; err != nil {
		respondError(c, err)
		return
	}
	listing := model.BusinessCareerListing{
		AlumniProfileID: profile.ID,
		ListingType:     request.ListingType,
		Name:            strings.TrimSpace(request.Name),
		Category:        strings.TrimSpace(request.Category),
		Description:     strings.TrimSpace(request.Description),
		ContactURL:      strings.TrimSpace(request.ContactURL),
		Status:          "pending",
	}
	if err := h.db.Create(&listing).Error; err != nil {
		respondError(c, err)
		return
	}
	rs.SuccessResponse(c, listing, http.StatusCreated)
}

func (h *Handler) RegisterForEvent(c *gin.Context) {
	account, err := util.GetAccountContext(c, constants.ACCOUNT_STUDENT)
	if err != nil {
		respondError(c, err)
		return
	}
	eventID, ok := pathID(c, "id")
	if !ok {
		return
	}
	var profile model.AlumniProfile
	if err := h.db.Where("account_id = ? AND status = ?", account.ID, "approved").First(&profile).Error; err != nil {
		respondError(c, err)
		return
	}
	var event model.AlumniEvent
	if err := h.db.Where("id = ? AND status = ?", eventID, "published").First(&event).Error; err != nil {
		respondError(c, err)
		return
	}
	registration := model.EventRegistration{EventID: event.ID, AlumniProfileID: profile.ID, Status: "registered"}
	if err := h.db.Where("event_id = ? AND alumni_profile_id = ?", event.ID, profile.ID).FirstOrCreate(&registration).Error; err != nil {
		respondError(c, err)
		return
	}
	rs.SuccessResponse(c, registration, http.StatusCreated)
}

func (h *Handler) MySummary(c *gin.Context) {
	account, err := util.GetAccountContext(c, constants.ACCOUNT_STUDENT)
	if err != nil {
		respondError(c, err)
		return
	}
	var profile model.AlumniProfile
	if err := h.db.Where("account_id = ?", account.ID).First(&profile).Error; err != nil {
		respondError(c, err)
		return
	}
	var registrations, listings int64
	h.db.Model(&model.EventRegistration{}).Where("alumni_profile_id = ? AND status = ?", profile.ID, "registered").Count(&registrations)
	h.db.Model(&model.BusinessCareerListing{}).Where("alumni_profile_id = ? AND status = ?", profile.ID, "approved").Count(&listings)
	c.JSON(http.StatusOK, gin.H{"message": "success", "content": gin.H{"profile": profile, "events_registered": registrations, "active_listings": listings}})
}

func (h *Handler) AdminStats(c *gin.Context) {
	var stats struct {
		Alumni             int64 `json:"alumni"`
		PendingProfiles    int64 `json:"pending_profiles"`
		PublishedNews      int64 `json:"published_news"`
		UpcomingEvents     int64 `json:"upcoming_events"`
		PendingListings    int64 `json:"pending_listings"`
		NewContactMessages int64 `json:"new_contact_messages"`
	}
	h.db.Model(&model.AlumniProfile{}).Where("status = ?", "approved").Count(&stats.Alumni)
	h.db.Model(&model.AlumniProfile{}).Where("status = ?", "pending").Count(&stats.PendingProfiles)
	h.db.Model(&model.NewsArticle{}).Where("status = ?", "published").Count(&stats.PublishedNews)
	h.db.Model(&model.AlumniEvent{}).Where("status = ? AND starts_at >= ?", "published", time.Now()).Count(&stats.UpcomingEvents)
	h.db.Model(&model.BusinessCareerListing{}).Where("status = ?", "pending").Count(&stats.PendingListings)
	h.db.Model(&model.ContactMessage{}).Where("status = ?", "new").Count(&stats.NewContactMessages)
	rs.SuccessResponse(c, stats)
}

func (h *Handler) AdminProfiles(c *gin.Context) {
	var profiles []model.AlumniProfile
	query := h.db.Preload("Account")
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Order("created_at DESC").Find(&profiles).Error; err != nil {
		respondError(c, err)
		return
	}
	items := make([]reviewedProfile, 0, len(profiles))
	for _, profile := range profiles {
		items = append(items, toReviewedProfile(profile))
	}
	rs.SuccessResponse(c, items)
}

func (h *Handler) ReviewProfile(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	var request struct {
		Status string `json:"status" binding:"required,oneof=approved rejected"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, err)
		return
	}
	result := h.db.Model(&model.AlumniProfile{}).Where("id = ?", id).Update("status", request.Status)
	if result.Error != nil {
		respondError(c, result.Error)
		return
	}
	if result.RowsAffected == 0 {
		respondError(c, gorm.ErrRecordNotFound)
		return
	}
	rs.SuccessResponse(c, gin.H{"status": request.Status})
}

func (h *Handler) AdminListings(c *gin.Context) {
	var listings []model.BusinessCareerListing
	query := h.db.Preload("AlumniProfile.Account")
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Order("created_at DESC").Find(&listings).Error; err != nil {
		respondError(c, err)
		return
	}
	items := make([]gin.H, 0, len(listings))
	for _, listing := range listings {
		owner := toPublicAlumni(listing.AlumniProfile)
		items = append(items, gin.H{"id": listing.ID, "listing_type": listing.ListingType, "name": listing.Name, "category": listing.Category, "description": listing.Description, "contact_url": listing.ContactURL, "status": listing.Status, "owner": owner, "reviewed_by": listing.ReviewedBy, "reviewed_at": listing.ReviewedAt})
	}
	rs.SuccessResponse(c, items)
}

func (h *Handler) ReviewListing(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	admin, err := util.GetAccountContext(c, constants.ACCOUNT_ADMIN)
	if err != nil {
		respondError(c, err)
		return
	}
	var request struct {
		Status string `json:"status" binding:"required,oneof=approved rejected"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, err)
		return
	}
	now := time.Now()
	result := h.db.Model(&model.BusinessCareerListing{}).Where("id = ?", id).Updates(map[string]interface{}{"status": request.Status, "reviewed_by": admin.ID, "reviewed_at": now})
	if result.Error != nil {
		respondError(c, result.Error)
		return
	}
	if result.RowsAffected == 0 {
		respondError(c, gorm.ErrRecordNotFound)
		return
	}
	rs.SuccessResponse(c, gin.H{"status": request.Status})
}

func (h *Handler) AdminNews(c *gin.Context) {
	var articles []model.NewsArticle
	query := h.db.Model(&model.NewsArticle{})
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Order("updated_at DESC").Find(&articles).Error; err != nil {
		respondError(c, err)
		return
	}
	rs.SuccessResponse(c, articles)
}

func (h *Handler) SaveNews(c *gin.Context) {
	admin, err := util.GetAccountContext(c, constants.ACCOUNT_ADMIN)
	if err != nil {
		respondError(c, err)
		return
	}
	request, ok := bind[articleRequest](c)
	if !ok {
		return
	}
	if request.Status == "" {
		request.Status = "draft"
	}
	if request.Status != "draft" && request.Status != "published" && request.Status != "archived" {
		respondError(c, errors.New("invalid article status"))
		return
	}
	article := model.NewsArticle{AuthorID: admin.ID, Title: strings.TrimSpace(request.Title), Slug: strings.TrimSpace(request.Slug), Category: strings.TrimSpace(request.Category), Excerpt: strings.TrimSpace(request.Excerpt), Body: strings.TrimSpace(request.Body), Status: request.Status}
	if request.Status == "published" {
		now := time.Now()
		article.PublishedAt = &now
	}
	if err := h.db.Create(&article).Error; err != nil {
		respondError(c, err)
		return
	}
	rs.SuccessResponse(c, article, http.StatusCreated)
}

func (h *Handler) UpdateNews(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	request, ok := bind[articleRequest](c)
	if !ok {
		return
	}
	if request.Status == "" {
		request.Status = "draft"
	}
	if request.Status != "draft" && request.Status != "published" && request.Status != "archived" {
		respondError(c, errors.New("invalid article status"))
		return
	}
	updates := map[string]interface{}{"title": strings.TrimSpace(request.Title), "slug": strings.TrimSpace(request.Slug), "category": strings.TrimSpace(request.Category), "excerpt": strings.TrimSpace(request.Excerpt), "body": strings.TrimSpace(request.Body), "status": request.Status}
	if request.Status == "published" {
		updates["published_at"] = time.Now()
	}
	result := h.db.Model(&model.NewsArticle{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		respondError(c, result.Error)
		return
	}
	if result.RowsAffected == 0 {
		respondError(c, gorm.ErrRecordNotFound)
		return
	}
	rs.SuccessResponse(c, gin.H{"status": "updated"})
}

func (h *Handler) DeleteNews(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	result := h.db.Delete(&model.NewsArticle{}, "id = ?", id)
	if result.Error != nil {
		respondError(c, result.Error)
		return
	}
	if result.RowsAffected == 0 {
		respondError(c, gorm.ErrRecordNotFound)
		return
	}
	rs.SuccessResponse(c, nil)
}

func (h *Handler) AdminEvents(c *gin.Context) {
	var events []model.AlumniEvent
	query := h.db.Model(&model.AlumniEvent{})
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Order("starts_at ASC").Find(&events).Error; err != nil {
		respondError(c, err)
		return
	}
	rs.SuccessResponse(c, events)
}

func (h *Handler) SaveEvent(c *gin.Context) {
	admin, err := util.GetAccountContext(c, constants.ACCOUNT_ADMIN)
	if err != nil {
		respondError(c, err)
		return
	}
	request, ok := bind[eventRequest](c)
	if !ok {
		return
	}
	if request.Status == "" {
		request.Status = "draft"
	}
	if request.Status != "draft" && request.Status != "published" && request.Status != "archived" {
		respondError(c, errors.New("invalid event status"))
		return
	}
	event := model.AlumniEvent{AuthorID: admin.ID, Title: strings.TrimSpace(request.Title), Description: strings.TrimSpace(request.Description), Category: strings.TrimSpace(request.Category), Venue: strings.TrimSpace(request.Venue), StartsAt: request.StartsAt, EndsAt: request.EndsAt, RegistrationDeadline: request.RegistrationDeadline, Status: request.Status}
	if err := h.db.Create(&event).Error; err != nil {
		respondError(c, err)
		return
	}
	rs.SuccessResponse(c, event, http.StatusCreated)
}

func (h *Handler) UpdateEvent(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	request, ok := bind[eventRequest](c)
	if !ok {
		return
	}
	if request.Status == "" {
		request.Status = "draft"
	}
	if request.Status != "draft" && request.Status != "published" && request.Status != "archived" {
		respondError(c, errors.New("invalid event status"))
		return
	}
	updates := map[string]interface{}{"title": strings.TrimSpace(request.Title), "description": strings.TrimSpace(request.Description), "category": strings.TrimSpace(request.Category), "venue": strings.TrimSpace(request.Venue), "starts_at": request.StartsAt, "ends_at": request.EndsAt, "registration_deadline": request.RegistrationDeadline, "status": request.Status}
	result := h.db.Model(&model.AlumniEvent{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		respondError(c, result.Error)
		return
	}
	if result.RowsAffected == 0 {
		respondError(c, gorm.ErrRecordNotFound)
		return
	}
	rs.SuccessResponse(c, gin.H{"status": "updated"})
}

func (h *Handler) DeleteEvent(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	result := h.db.Delete(&model.AlumniEvent{}, "id = ?", id)
	if result.Error != nil {
		respondError(c, result.Error)
		return
	}
	if result.RowsAffected == 0 {
		respondError(c, gorm.ErrRecordNotFound)
		return
	}
	rs.SuccessResponse(c, nil)
}

func (h *Handler) AdminContacts(c *gin.Context) {
	var messages []model.ContactMessage
	query := h.db.Model(&model.ContactMessage{})
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Order("created_at DESC").Find(&messages).Error; err != nil {
		respondError(c, err)
		return
	}
	rs.SuccessResponse(c, messages)
}

func (h *Handler) UpdateContactStatus(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	var request struct {
		Status string `json:"status" binding:"required,oneof=read closed"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, err)
		return
	}
	result := h.db.Model(&model.ContactMessage{}).Where("id = ?", id).Update("status", request.Status)
	if result.Error != nil {
		respondError(c, result.Error)
		return
	}
	if result.RowsAffected == 0 {
		respondError(c, gorm.ErrRecordNotFound)
		return
	}
	rs.SuccessResponse(c, gin.H{"status": request.Status})
}
