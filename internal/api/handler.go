package api

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"

	contract "github.com/neiln3121/explore-service/explore"
	"github.com/neiln3121/explore-service/internal/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//go:generate go run github.com/vektra/mockery/v2@v2.50.4 --with-expecter --exported --name=Store
type Store interface {
	PutDecision(ctx context.Context, recipient_id, actor_id string, liked bool) error
	PutMutualDecisions(ctx context.Context, recipient_id, actor_id string, liked, mutuallyLiked bool) error

	GetLikedDecisions(ctx context.Context, recipientID string, liked bool, token *uint64, limit *uint32) ([]*storage.Liker, error)
	GetNewLikedDecisions(ctx context.Context, recipientID string, liked bool, token *uint64, limit *uint32) ([]*storage.Liker, error)
	GetLikedDecisionsCount(ctx context.Context, recipientID string, liked bool) (int, error)
	GetLikedDecision(ctx context.Context, recipientID, actorID string) (bool, error)
}

// ExporeAPI is the implementation of the GRPC server
type ExploreAPI struct {
	repository Store
	contract.UnimplementedExploreAPIServer
}

func New(repository Store) *ExploreAPI {
	return &ExploreAPI{
		repository: repository,
	}
}

func (e *ExploreAPI) ListLikedYou(ctx context.Context, req *contract.ListLikedYouRequest) (*contract.ListLikedYouResponse, error) {
	token, err := validateListLikedYouRequest(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	likers, err := e.repository.GetLikedDecisions(ctx, req.RecipientUserId, true, token, req.PaginationLimit)
	if err != nil {
		log.Printf("Internal error on GetLikedDecisions call: %v", err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get recipient likes, %s", err))
	}

	responseLikers := make([]*contract.ListLikedYouResponse_Liker, len(likers))
	for index, liker := range likers {
		responseLikers[index] = &contract.ListLikedYouResponse_Liker{
			ActorId:   liker.ActorID,
			UpdatedAt: uint64(liker.UpdatedAt),
		}
	}

	// Use ID from last liker as token
	var nextToken *string
	if req.PaginationLimit != nil {
		index := len(likers)
		if index > 0 {
			index--
			uintToken := strconv.FormatUint(likers[index].ID, 10)
			nextToken = &uintToken
		}
	}
	return &contract.ListLikedYouResponse{
		Likers:              responseLikers,
		NextPaginationToken: nextToken,
	}, nil
}

func (e *ExploreAPI) ListNewLikedYou(ctx context.Context, req *contract.ListLikedYouRequest) (*contract.ListLikedYouResponse, error) {
	token, err := validateListLikedYouRequest(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	likers, err := e.repository.GetNewLikedDecisions(ctx, req.RecipientUserId, true, token, req.PaginationLimit)
	if err != nil {
		log.Printf("Internal error on GetNewLikedDecisions call: %v", err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get new recipient likes, %s", err))
	}

	responseLikers := make([]*contract.ListLikedYouResponse_Liker, len(likers))
	for index, liker := range likers {
		responseLikers[index] = &contract.ListLikedYouResponse_Liker{
			ActorId:   liker.ActorID,
			UpdatedAt: uint64(liker.UpdatedAt),
		}
	}

	// Use ID from last liker as token
	var nextToken *string
	if req.PaginationLimit != nil {
		index := len(likers)
		if index > 0 {
			index--
			uintToken := strconv.FormatUint(likers[index].ID, 10)
			nextToken = &uintToken
		}
	}
	return &contract.ListLikedYouResponse{
		Likers:              responseLikers,
		NextPaginationToken: nextToken,
	}, nil
}

func (e *ExploreAPI) CountLikedYou(ctx context.Context, req *contract.CountLikedYouRequest) (*contract.CountLikedYouResponse, error) {
	if req.RecipientUserId == "" {
		return nil, status.Error(codes.InvalidArgument, "empty recipient ID")
	}
	res, err := e.repository.GetLikedDecisionsCount(ctx, req.RecipientUserId, true)
	if err != nil {
		log.Printf("Internal error on GetLikedDecisionsCount call: %v", err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get recipient liked count, %s", err))
	}
	return &contract.CountLikedYouResponse{
		Count: uint64(res),
	}, nil
}

func (e *ExploreAPI) PutDecision(ctx context.Context, req *contract.PutDecisionRequest) (*contract.PutDecisionResponse, error) {
	firstToLike := false
	// Determine if there is a decision already from the recipient
	recipientLiked, err := e.repository.GetLikedDecision(ctx, req.ActorUserId, req.RecipientUserId)
	if err != nil {
		if errors.Is(err, storage.ErrDecisionNotFound) {
			firstToLike = true
		} else {
			log.Printf("Internal error on GetLikedDecision call: %v", err)
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed to determine mutual decision, %s", err))
		}
	}

	// If there is no mutual decision, just put a single decision for the recipient
	if firstToLike {
		err = e.repository.PutDecision(ctx, req.RecipientUserId, req.ActorUserId, req.LikedRecipient)
		if err != nil {
			log.Printf("Internal error on PutDecision call: %v", err)
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update decision, %s", err))
		}
	} else {
		// If we have a mutual decision then put the decision for recipient with mutually liked set, and update actor decision for mutually liked
		err = e.repository.PutMutualDecisions(ctx, req.RecipientUserId, req.ActorUserId, req.LikedRecipient, recipientLiked)
		if err != nil {
			log.Printf("Internal error on PutMutualDecisions call: %v", err)
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update mutual decision, %s", err))
		}
	}

	return &contract.PutDecisionResponse{
		MutualLikes: recipientLiked == req.LikedRecipient,
	}, nil
}

func validateListLikedYouRequest(req *contract.ListLikedYouRequest) (*uint64, error) {
	if req.RecipientUserId == "" {
		return nil, fmt.Errorf("empty recipient ID")
	}
	if req.PaginationToken != nil {
		if *req.PaginationToken == "" {
			return nil, fmt.Errorf("empty pagination token")
		}
		res, err := strconv.ParseUint(*req.PaginationToken, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid pagination token: %w", err)
		}
		return &res, nil
	}

	return nil, nil
}
