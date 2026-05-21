package service

import (
	"rare_backend/internal/module/community/domain"
	"rare_backend/internal/module/community/repo"
)

type CommentService struct {
	postRepo    *repo.PostRepo
	commentRepo *repo.CommentRepo
}

func NewCommentService(postR *repo.PostRepo, commentR *repo.CommentRepo) *CommentService {
	return &CommentService{postRepo: postR, commentRepo: commentR}
}

func (s *CommentService) Create(in domain.CreateCommentInput) (*domain.CreateCommentResult, error) {
	ok, err := s.postRepo.ExistsActive(in.PostID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, domain.ErrPostNotFound
	}

	parentID := int64(0)
	rootID := int64(0)
	if in.ParentID != nil && *in.ParentID > 0 {
		parentID = *in.ParentID
		parent, err := s.commentRepo.FindParentOnPost(parentID, in.PostID)
		if err != nil {
			if s.commentRepo.ErrNoRows(err) {
				return nil, domain.ErrParentComment
			}
			return nil, err
		}
		if parent.RootID == 0 {
			rootID = parentID
		} else {
			rootID = parent.RootID
		}
	}
	return s.commentRepo.CreateWithTx(in, parentID, rootID)
}

func (s *CommentService) Update(commentID int64, userID int64, content string, isAdmin bool) error {
	ownerID, err := s.commentRepo.GetOwnerActive(commentID)
	if err != nil {
		if s.commentRepo.ErrNoRows(err) {
			return domain.ErrCommentNotFound
		}
		return err
	}
	if !isAdmin && ownerID != userID {
		return domain.ErrForbidden
	}
	return s.commentRepo.UpdateContent(commentID, content)
}

func (s *CommentService) Delete(commentID, userID int64, isAdmin bool) error {
	ownerID, postID, err := s.commentRepo.GetOwnerAndPost(commentID)
	if err != nil {
		if s.commentRepo.ErrNoRows(err) {
			return domain.ErrCommentNotFound
		}
		return err
	}
	if !isAdmin && ownerID != userID {
		return domain.ErrForbidden
	}
	return s.commentRepo.SoftDeleteWithDecrement(commentID, postID)
}
