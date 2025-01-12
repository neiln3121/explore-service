package api_test

import (
	"context"
	"errors"
	"testing"

	contract "github.com/neiln3121/explore-service/explore"
	"github.com/neiln3121/explore-service/internal/api"
	"github.com/neiln3121/explore-service/internal/api/mocks"
	"github.com/neiln3121/explore-service/internal/storage"
	"github.com/stretchr/testify/assert"
)

var (
	errorDB                    = errors.New("db error")
	testPaginationLimit        = uint32(1)
	testValidPaginationToken   = "2"
	testUintPaginationToken    = uint64(2)
	testInvalidPaginationToken = "a"
)

type mockCall struct {
	recipientID     string
	paginationToken *uint64
	paginationLimit *uint32
}

func Test_GetCountLikedYou(t *testing.T) {
	ctx := context.Background()

	store := mocks.NewStore(t)
	api := api.New(store)

	testCases := []struct {
		description    string
		request        *contract.CountLikedYouRequest
		noMockCall     bool
		mockResponse   int
		mockError      error
		expectedResult uint64
		expectedError  string
	}{
		{
			description: "valid",
			request: &contract.CountLikedYouRequest{
				RecipientUserId: "1",
			},
			mockResponse:   4,
			expectedResult: 4,
		},
		{
			description: "db error",
			request: &contract.CountLikedYouRequest{
				RecipientUserId: "1",
			},
			mockError:     errorDB,
			expectedError: "rpc error: code = Internal desc = failed to get recipient liked count, db error",
		},
		{
			description: "invalid request",
			request: &contract.CountLikedYouRequest{
				RecipientUserId: "",
			},
			noMockCall:    true,
			expectedError: "rpc error: code = InvalidArgument desc = empty recipient ID",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			if !tc.noMockCall {
				store.EXPECT().GetLikedDecisionsCount(ctx, "1", true).Return(tc.mockResponse, tc.mockError).Once()
			}
			res, err := api.CountLikedYou(ctx, tc.request)

			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResult, res.Count)
			} else {
				assert.EqualError(t, err, tc.expectedError)
				assert.Nil(t, res)
			}

		})
	}
}

func Test_GetListLikedYou(t *testing.T) {
	ctx := context.Background()

	store := mocks.NewStore(t)
	api := api.New(store)

	testCases := []struct {
		description             string
		request                 *contract.ListLikedYouRequest
		mockCall                *mockCall
		mockResponse            []*storage.Liker
		mockError               error
		expectedResult          []*contract.ListLikedYouResponse_Liker
		expectedPaginationToken *string
		expectedError           string
	}{
		{
			description: "valid",
			request: &contract.ListLikedYouRequest{
				RecipientUserId: "1",
			},
			mockCall: &mockCall{
				recipientID: "1",
			},
			mockResponse: []*storage.Liker{
				{
					ID:        1,
					ActorID:   "actor-1",
					UpdatedAt: 1,
				},
				{
					ID:        2,
					ActorID:   "actor-2",
					UpdatedAt: 2,
				},
			},
			expectedResult: []*contract.ListLikedYouResponse_Liker{
				{
					ActorId:   "actor-1",
					UpdatedAt: 1,
				},
				{
					ActorId:   "actor-2",
					UpdatedAt: 2,
				},
			},
		},
		{
			description: "db error",
			request: &contract.ListLikedYouRequest{
				RecipientUserId: "1",
			},
			mockCall: &mockCall{
				recipientID: "1",
			},
			mockError:     errorDB,
			expectedError: "rpc error: code = Internal desc = failed to get recipient likes, db error",
		},
		{
			description: "valid pagination token",
			request: &contract.ListLikedYouRequest{
				RecipientUserId: "1",
				PaginationToken: &testValidPaginationToken,
				PaginationLimit: &testPaginationLimit,
			},
			mockCall: &mockCall{
				recipientID:     "1",
				paginationToken: &testUintPaginationToken,
				paginationLimit: &testPaginationLimit,
			},
			mockResponse: []*storage.Liker{
				{
					ID:        1,
					ActorID:   "actor-1",
					UpdatedAt: 1,
				},
				{
					ID:        2,
					ActorID:   "actor-2",
					UpdatedAt: 2,
				},
			},
			expectedResult: []*contract.ListLikedYouResponse_Liker{
				{
					ActorId:   "actor-1",
					UpdatedAt: 1,
				},
				{
					ActorId:   "actor-2",
					UpdatedAt: 2,
				},
			},
			expectedPaginationToken: &testValidPaginationToken,
		},
		{
			description: "invalid pagination token",
			request: &contract.ListLikedYouRequest{
				RecipientUserId: "1",
				PaginationToken: &testInvalidPaginationToken,
				PaginationLimit: &testPaginationLimit,
			},
			expectedError: "rpc error: code = InvalidArgument desc = invalid pagination token: strconv.ParseUint: parsing \"a\": invalid syntax",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			if tc.mockCall != nil {
				store.EXPECT().GetLikedDecisions(
					ctx,
					tc.mockCall.recipientID,
					true,
					tc.mockCall.paginationToken,
					tc.mockCall.paginationLimit,
				).Return(tc.mockResponse, tc.mockError).Once()
			}
			res, err := api.ListLikedYou(ctx, tc.request)

			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResult, res.Likers)
				assert.Equal(t, tc.expectedPaginationToken, res.NextPaginationToken)
			} else {
				assert.EqualError(t, err, tc.expectedError)
				assert.Nil(t, res)
			}

		})
	}
}

func Test_GetListNewLikedYou(t *testing.T) {
	ctx := context.Background()

	store := mocks.NewStore(t)
	api := api.New(store)

	testCases := []struct {
		description             string
		request                 *contract.ListLikedYouRequest
		mockCall                *mockCall
		mockResponse            []*storage.Liker
		mockError               error
		expectedResult          []*contract.ListLikedYouResponse_Liker
		expectedPaginationToken *string
		expectedError           string
	}{
		{
			description: "valid",
			request: &contract.ListLikedYouRequest{
				RecipientUserId: "1",
			},
			mockCall: &mockCall{
				recipientID: "1",
			},
			mockResponse: []*storage.Liker{
				{
					ID:        1,
					ActorID:   "actor-1",
					UpdatedAt: 1,
				},
				{
					ID:        2,
					ActorID:   "actor-2",
					UpdatedAt: 2,
				},
			},
			expectedResult: []*contract.ListLikedYouResponse_Liker{
				{
					ActorId:   "actor-1",
					UpdatedAt: 1,
				},
				{
					ActorId:   "actor-2",
					UpdatedAt: 2,
				},
			},
		},
		{
			description: "db error",
			request: &contract.ListLikedYouRequest{
				RecipientUserId: "1",
			},
			mockCall: &mockCall{
				recipientID: "1",
			},
			mockError:     errorDB,
			expectedError: "rpc error: code = Internal desc = failed to get new recipient likes, db error",
		},
		{
			description: "valid pagination token",
			request: &contract.ListLikedYouRequest{
				RecipientUserId: "1",
				PaginationToken: &testValidPaginationToken,
				PaginationLimit: &testPaginationLimit,
			},
			mockCall: &mockCall{
				recipientID:     "1",
				paginationToken: &testUintPaginationToken,
				paginationLimit: &testPaginationLimit,
			},
			mockResponse: []*storage.Liker{
				{
					ID:        1,
					ActorID:   "actor-1",
					UpdatedAt: 1,
				},
				{
					ID:        2,
					ActorID:   "actor-2",
					UpdatedAt: 2,
				},
			},
			expectedResult: []*contract.ListLikedYouResponse_Liker{
				{
					ActorId:   "actor-1",
					UpdatedAt: 1,
				},
				{
					ActorId:   "actor-2",
					UpdatedAt: 2,
				},
			},
			expectedPaginationToken: &testValidPaginationToken,
		},
		{
			description: "invalid pagination token",
			request: &contract.ListLikedYouRequest{
				RecipientUserId: "1",
				PaginationToken: &testInvalidPaginationToken,
				PaginationLimit: &testPaginationLimit,
			},
			expectedError: "rpc error: code = InvalidArgument desc = invalid pagination token: strconv.ParseUint: parsing \"a\": invalid syntax",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			if tc.mockCall != nil {
				store.EXPECT().GetNewLikedDecisions(
					ctx,
					tc.mockCall.recipientID,
					true,
					tc.mockCall.paginationToken,
					tc.mockCall.paginationLimit,
				).Return(tc.mockResponse, tc.mockError).Once()
			}

			res, err := api.ListNewLikedYou(ctx, tc.request)

			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResult, res.Likers)
				assert.Equal(t, tc.expectedPaginationToken, res.NextPaginationToken)
			} else {
				assert.EqualError(t, err, tc.expectedError)
				assert.Nil(t, res)
			}

		})
	}
}

func Test_PutDecision(t *testing.T) {
	type mockLikedDecisionResponse struct {
		liked bool
		err   error
	}

	ctx := context.Background()

	store := mocks.NewStore(t)
	api := api.New(store)

	testCases := []struct {
		description string
		request     *contract.PutDecisionRequest

		mockGetLikedDecisionResponse mockLikedDecisionResponse
		shouldPutDecision            bool
		shouldPutMutualDecision      bool
		mockError                    error
		expectedResult               bool
		expectedError                string
	}{
		{
			description: "valid - first like",
			request: &contract.PutDecisionRequest{
				RecipientUserId: "recipient-1",
				ActorUserId:     "actor-1",
				LikedRecipient:  true,
			},
			mockGetLikedDecisionResponse: mockLikedDecisionResponse{
				err: storage.ErrDecisionNotFound,
			},
			shouldPutDecision: true,
			expectedResult:    false,
		},
		{
			description: "db error - first like",
			request: &contract.PutDecisionRequest{
				RecipientUserId: "recipient-1",
				ActorUserId:     "actor-1",
				LikedRecipient:  true,
			},
			mockGetLikedDecisionResponse: mockLikedDecisionResponse{
				err: errorDB,
			},
			expectedError: "rpc error: code = Internal desc = failed to determine mutual decision, db error",
		},
		{
			description: "db error - put decision",
			request: &contract.PutDecisionRequest{
				RecipientUserId: "recipient-1",
				ActorUserId:     "actor-1",
				LikedRecipient:  true,
			},
			mockGetLikedDecisionResponse: mockLikedDecisionResponse{
				err: storage.ErrDecisionNotFound,
			},
			shouldPutDecision: true,
			mockError:         errorDB,
			expectedError:     "rpc error: code = Internal desc = failed to update decision, db error",
		},
		{
			description: "valid - mutual likes",
			request: &contract.PutDecisionRequest{
				RecipientUserId: "recipient-1",
				ActorUserId:     "actor-1",
				LikedRecipient:  true,
			},
			mockGetLikedDecisionResponse: mockLikedDecisionResponse{
				liked: true,
			},
			shouldPutMutualDecision: true,
			expectedResult:          true,
		},
		{
			description: "valid - recipient doesn't like back",
			request: &contract.PutDecisionRequest{
				RecipientUserId: "recipient-1",
				ActorUserId:     "actor-1",
				LikedRecipient:  true,
			},
			mockGetLikedDecisionResponse: mockLikedDecisionResponse{
				liked: false,
			},
			shouldPutMutualDecision: true,
			expectedResult:          false,
		},
		{
			description: "db error - put mutual decision",
			request: &contract.PutDecisionRequest{
				RecipientUserId: "recipient-1",
				ActorUserId:     "actor-1",
				LikedRecipient:  true,
			},
			mockGetLikedDecisionResponse: mockLikedDecisionResponse{
				liked: true,
			},
			shouldPutMutualDecision: true,
			mockError:               errorDB,
			expectedError:           "rpc error: code = Internal desc = failed to update mutual decision, db error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			store.EXPECT().GetLikedDecision(
				ctx,
				tc.request.ActorUserId,
				tc.request.RecipientUserId,
			).Return(tc.mockGetLikedDecisionResponse.liked, tc.mockGetLikedDecisionResponse.err).Once()

			if tc.shouldPutDecision {
				store.EXPECT().PutDecision(
					ctx,
					tc.request.RecipientUserId,
					tc.request.ActorUserId,
					tc.request.LikedRecipient,
				).Return(tc.mockError).Once()
			}

			if tc.shouldPutMutualDecision {
				store.EXPECT().PutMutualDecisions(
					ctx,
					tc.request.RecipientUserId,
					tc.request.ActorUserId,
					tc.request.LikedRecipient,
					tc.mockGetLikedDecisionResponse.liked,
				).Return(tc.mockError).Once()
			}

			res, err := api.PutDecision(ctx, tc.request)

			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResult, res.MutualLikes)
			} else {
				assert.EqualError(t, err, tc.expectedError)
				assert.Nil(t, res)
			}

		})
	}
}
