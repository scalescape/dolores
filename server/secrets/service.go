package secrets

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/scalescape/dolores/server/cerr"
	"github.com/scalescape/dolores/server/cloud"
	"github.com/scalescape/dolores/server/org"
	"github.com/scalescape/dolores/server/platform"
)

type projFetcher interface {
	FetchProject(ctx context.Context, oid string, env string) (platform.Project, error)
}
type Service struct {
	proj projFetcher
	Store
}

type Recipient struct {
	PublicKey string `json:"public_key"`
}

func (s Service) getStorageClient(ctx context.Context, proj platform.Project) (cloud.StorageClient, error) {
	cfg := &cloud.Config{OrgID: proj.OrgID, ProjectID: proj.ID, Credentials: []byte(proj.Credentials), Platform: cloud.Platform(proj.Platform)}
	if proj.Region.Valid {
		cfg.Region = proj.Region.String
	}
	sc, err := cloud.NewStorageClient(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %w", err)
	}
	return sc, nil
}

func (s Service) ListRecipients(ctx context.Context, env string) ([]Recipient, error) {
	oid, ok := ctx.Value(org.IDKey).(string)
	if !ok || oid == "" {
		return nil, fmt.Errorf("empty org id: %w", cerr.ErrInvalidOrg)
	}
	log.Trace().Msgf("fetching user public keys, env: %s org_id: %s", env, oid)
	result, err := s.Store.ListUsersPublicKeys(ctx, env, oid)
	if err != nil {
		return nil, err
	}
	recps := make([]Recipient, len(result))
	for i, uk := range result {
		recps[i].PublicKey = uk.PublicKey
	}
	return recps, nil
}

func (s Service) ListSecret(ctx context.Context, req listRequest) ([]Secret, error) {
	var secs []Secret
	proj, err := s.proj.FetchProject(ctx, req.orgID, string(req.environment))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch project: %w", err)
	}
	log.Trace().Msgf("listing objects for org: %s, project: %s bucket: %s", req.orgID, proj.ID, proj.Bucket)
	sc, err := s.getStorageClient(ctx, proj)
	if err != nil {
		return nil, err
	}
	objs, err := sc.ListObject(ctx, proj.Bucket, "secrets")
	if err != nil {
		return nil, err
	}
	for _, obj := range objs {
		if !strings.HasSuffix(obj.Name, ".key") && !strings.HasSuffix(obj.Name, "/") {
			secs = append(secs, Secret{
				Name: obj.Name, CreatedAt: obj.CreatedAt,
				UpdatedAt: obj.UpdatedAt, Location: fmt.Sprintf("%s/%s", obj.Bucket, obj.Name),
			})
		}
	}
	return secs, nil
}

func (s Service) UploadSecret(ctx context.Context, req uploadRequest) error {
	oid, ok := ctx.Value(org.IDKey).(string)
	if !ok || oid == "" {
		return cerr.ErrInvalidOrgID
	}
	proj, err := s.proj.FetchProject(ctx, oid, req.Environment)
	if err != nil {
		return fmt.Errorf("failed to fetch project: %w", err)
	}
	sc, err := s.getStorageClient(ctx, proj)
	if err != nil {
		return err
	}
	name := fmt.Sprintf("secrets/%s", req.Name)
	err = sc.WriteToObject(ctx, proj.Bucket, name, req.decodedData)
	if err != nil {
		return fmt.Errorf("failed to write to gcs: %w", err)
	}
	sec := Secret{ID: uuid.New().String(), Name: req.Name, ProjectID: proj.ID, Location: name, CreatedAt: time.Now().UTC()}
	if err := s.Store.SaveSecret(ctx, sec); err != nil {
		return err
	}
	return nil
}

func (s Service) FetchSecret(ctx context.Context, req fetchRequest) (string, error) {
	proj, err := s.proj.FetchProject(ctx, req.orgID, string(req.Environment))
	if err != nil {
		return "", fmt.Errorf("failed to fetch project: %w", err)
	}
	//sec, err := s.Store.fetchSecret(ctx, req.Name, proj.ID)
	//if err != nil {
	//	return "", err
	//}
	sc, err := s.getStorageClient(ctx, proj)
	if err != nil {
		return "", err
	}
	location := fmt.Sprintf("secrets/%s", req.Name)
	data, err := sc.ReadObject(ctx, proj.Bucket, location)
	if err != nil {
		return "", fmt.Errorf("error reading object: %w", err)
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

func NewService(st Store, pj projFetcher) Service {
	return Service{Store: st, proj: pj}
}
