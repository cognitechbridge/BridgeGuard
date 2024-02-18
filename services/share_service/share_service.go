package share_service

import (
	"ctb-cli/keystore"
	"ctb-cli/repositories"
	"ctb-cli/services/object_service"
	"encoding/base64"
)

type Service struct {
	recipientRepository repositories.RecipientRepository
	linkRepository      *repositories.LinkRepository
	objectService       *object_service.Service
	keyStorer           keystore.KeyStorer
}

func (s *Service) ShareByEmail(regex string, email string) error {
	rec, _ := s.recipientRepository.GetRecipientByEmail(email)
	files, _ := s.linkRepository.ListIdsByRegex(regex)
	for _, fileId := range files {
		keyId, err := s.objectService.KetKeyIdByObjectId(fileId)
		if err != nil {
			return err
		}
		publicBytes, err := base64.RawStdEncoding.DecodeString(rec.Public)
		s.keyStorer.Share(keyId, publicBytes, rec.ClientId)
	}
	return nil
}

func NewService(
	recipientRepository repositories.RecipientRepository,
	keyStorer keystore.KeyStorer,
	linkRepository *repositories.LinkRepository,
	objectService *object_service.Service,
) *Service {
	return &Service{
		objectService:       objectService,
		recipientRepository: recipientRepository,
		keyStorer:           keyStorer,
		linkRepository:      linkRepository,
	}
}