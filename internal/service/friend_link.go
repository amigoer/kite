package service

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/amigoer/kite-blog/internal/model"
	"github.com/amigoer/kite-blog/internal/repo"
	"github.com/google/uuid"
)

var (
	ErrInvalidFriendLinkPayload = errors.New("invalid friend link payload")
	ErrDuplicateFriendLinkURL   = errors.New("duplicate friend link url")
)

type FriendLinkListParams struct {
	Page     int
	PageSize int
	Keyword  string
	IsActive *bool
}

type CreateFriendLinkInput struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Logo        string `json:"logo"`
	Sort        int    `json:"sort"`
	IsActive    *bool  `json:"is_active"`
}

type UpdateFriendLinkInput struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Logo        string `json:"logo"`
	Sort        int    `json:"sort"`
	IsActive    bool   `json:"is_active"`
}

type PatchFriendLinkInput struct {
	Name        *string `json:"name"`
	URL         *string `json:"url"`
	Description *string `json:"description"`
	Logo        *string `json:"logo"`
	Sort        *int    `json:"sort"`
	IsActive    *bool   `json:"is_active"`
}

type FriendLinkListResult struct {
	Items      []model.FriendLink `json:"items"`
	Pagination Pagination         `json:"pagination"`
}

type FriendLinkService struct {
	repo *repo.FriendLinkRepository
}

func NewFriendLinkService(repo *repo.FriendLinkRepository) *FriendLinkService {
	return &FriendLinkService{repo: repo}
}

func (s *FriendLinkService) List(params FriendLinkListParams) (*FriendLinkListResult, error) {
	return s.list(params, false)
}

func (s *FriendLinkService) ListPublic(params FriendLinkListParams) (*FriendLinkListResult, error) {
	return s.list(params, true)
}

func (s *FriendLinkService) list(params FriendLinkListParams, publicOnly bool) (*FriendLinkListResult, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("friend link service is unavailable")
	}

	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}
	if params.PageSize > 100 {
		params.PageSize = 100
	}

	items, total, err := s.repo.List(repo.FriendLinkListParams{
		Page:       params.Page,
		PageSize:   params.PageSize,
		Keyword:    strings.TrimSpace(params.Keyword),
		IsActive:   normalizeFriendLinkActiveFilter(params.IsActive, publicOnly),
		PublicOnly: publicOnly,
	})
	if err != nil {
		return nil, err
	}

	return &FriendLinkListResult{
		Items: items,
		Pagination: Pagination{
			Page:     params.Page,
			PageSize: params.PageSize,
			Total:    total,
		},
	}, nil
}

func (s *FriendLinkService) GetByID(id string) (*model.FriendLink, error) {
	parsedID, err := uuid.Parse(strings.TrimSpace(id))
	if err != nil {
		return nil, fmt.Errorf("%w: invalid friend link id", ErrInvalidFriendLinkPayload)
	}
	return s.repo.GetByID(parsedID)
}

func (s *FriendLinkService) GetPublicByID(id string) (*model.FriendLink, error) {
	parsedID, err := uuid.Parse(strings.TrimSpace(id))
	if err != nil {
		return nil, fmt.Errorf("%w: invalid friend link id", ErrInvalidFriendLinkPayload)
	}
	return s.repo.GetPublicByID(parsedID)
}

func (s *FriendLinkService) Create(input CreateFriendLinkInput) (*model.FriendLink, error) {
	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	link := &model.FriendLink{
		Name:        strings.TrimSpace(input.Name),
		URL:         strings.TrimSpace(input.URL),
		Description: strings.TrimSpace(input.Description),
		Logo:        strings.TrimSpace(input.Logo),
		Sort:        input.Sort,
		IsActive:    isActive,
	}

	if err := validateFriendLink(link); err != nil {
		return nil, err
	}
	if err := ensureFriendLinkURLAvailable(s.repo, link.URL, uuid.Nil); err != nil {
		return nil, err
	}
	if err := s.repo.Create(link); err != nil {
		return nil, err
	}
	return link, nil
}

func (s *FriendLinkService) Update(id string, input UpdateFriendLinkInput) (*model.FriendLink, error) {
	parsedID, err := uuid.Parse(strings.TrimSpace(id))
	if err != nil {
		return nil, fmt.Errorf("%w: invalid friend link id", ErrInvalidFriendLinkPayload)
	}

	link, err := s.repo.GetByID(parsedID)
	if err != nil {
		return nil, err
	}

	link.Name = strings.TrimSpace(input.Name)
	link.URL = strings.TrimSpace(input.URL)
	link.Description = strings.TrimSpace(input.Description)
	link.Logo = strings.TrimSpace(input.Logo)
	link.Sort = input.Sort
	link.IsActive = input.IsActive

	if err := validateFriendLink(link); err != nil {
		return nil, err
	}
	if err := ensureFriendLinkURLAvailable(s.repo, link.URL, link.ID); err != nil {
		return nil, err
	}
	if err := s.repo.Update(link); err != nil {
		return nil, err
	}
	return link, nil
}

func (s *FriendLinkService) Patch(id string, input PatchFriendLinkInput) (*model.FriendLink, error) {
	parsedID, err := uuid.Parse(strings.TrimSpace(id))
	if err != nil {
		return nil, fmt.Errorf("%w: invalid friend link id", ErrInvalidFriendLinkPayload)
	}

	link, err := s.repo.GetByID(parsedID)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		link.Name = strings.TrimSpace(*input.Name)
	}
	if input.URL != nil {
		link.URL = strings.TrimSpace(*input.URL)
	}
	if input.Description != nil {
		link.Description = strings.TrimSpace(*input.Description)
	}
	if input.Logo != nil {
		link.Logo = strings.TrimSpace(*input.Logo)
	}
	if input.Sort != nil {
		link.Sort = *input.Sort
	}
	if input.IsActive != nil {
		link.IsActive = *input.IsActive
	}

	if err := validateFriendLink(link); err != nil {
		return nil, err
	}
	if err := ensureFriendLinkURLAvailable(s.repo, link.URL, link.ID); err != nil {
		return nil, err
	}
	if err := s.repo.Update(link); err != nil {
		return nil, err
	}
	return link, nil
}

func (s *FriendLinkService) Delete(id string) error {
	parsedID, err := uuid.Parse(strings.TrimSpace(id))
	if err != nil {
		return fmt.Errorf("%w: invalid friend link id", ErrInvalidFriendLinkPayload)
	}
	return s.repo.Delete(parsedID)
}

func validateFriendLink(link *model.FriendLink) error {
	if link == nil {
		return fmt.Errorf("%w: friend link is required", ErrInvalidFriendLinkPayload)
	}
	if link.Name == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidFriendLinkPayload)
	}
	if link.URL == "" {
		return fmt.Errorf("%w: url is required", ErrInvalidFriendLinkPayload)
	}
	if err := validateURL(link.URL); err != nil {
		return err
	}
	return nil
}

func validateURL(rawURL string) error {
	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("%w: invalid url", ErrInvalidFriendLinkPayload)
	}
	return nil
}

func ensureFriendLinkURLAvailable(friendLinkRepo *repo.FriendLinkRepository, rawURL string, currentID uuid.UUID) error {
	existing, err := friendLinkRepo.GetByURL(rawURL)
	if err == nil && existing != nil && existing.ID != currentID {
		return ErrDuplicateFriendLinkURL
	}
	if err != nil && !errors.Is(err, repo.ErrFriendLinkNotFound) {
		return err
	}
	return nil
}

func normalizeFriendLinkActiveFilter(isActive *bool, publicOnly bool) *bool {
	if publicOnly {
		return nil
	}
	return isActive
}
