package filestorage

import (
	"context"
	"errors"
	"log/slog"
	"path"
	"runtime/debug"
	"strings"

	"github.com/sfperusacdev/identitysdk"
	"github.com/sfperusacdev/identitysdk/helpers/storage"
	"github.com/sfperusacdev/identitysdk/services"
	"github.com/user0608/goones/errs"
)

type GenericFileStorage struct {
	bridge       *services.ExternalBridgeService
	variableName string
}

var _ FileStorage = (*GenericFileStorage)(nil)

func NewGenericFileStorage(
	variableName string,
	bridge *services.ExternalBridgeService,
) *GenericFileStorage {
	return &GenericFileStorage{
		bridge:       bridge,
		variableName: strings.TrimSpace(variableName),
	}
}

func (uc *GenericFileStorage) newFileStorer(ctx context.Context) (storage.FileStorer, error) {
	if uc.variableName == "" {
		slog.Error(
			"nombre de variable no encontrado",
			"trace", string(debug.Stack()),
		)
		return nil, errs.InternalErrorDirect(errs.ErrInternal)
	}
	basePath, err := uc.bridge.ReadVariable(ctx, uc.variableName)
	if err != nil {
		return nil, errs.BadRequestf("'%s' es requerido pero no fue encontrado", uc.variableName)
	}
	if basePath == "" {
		return nil, errs.BadRequestf("valor faltante para '%s'", uc.variableName)
	}

	storagePath := path.Join(
		basePath,
		identitysdk.Empresa(ctx),
	)

	fileStorer, err := uc.bridge.CreateNewFileStorer(ctx, storagePath)
	if err != nil {
		slog.Error("failed to create file storer", "path", storagePath, "error", err)
		return nil, err
	}
	return fileStorer, nil
}

// Store implements services.FileStorage.
func (uc *GenericFileStorage) Store(ctx context.Context, fileName string, data []byte) error {
	if strings.TrimSpace(fileName) == "" {
		return errs.BadRequestDirect("fileName es requerido")
	}
	storer, err := uc.newFileStorer(ctx)
	if err != nil {
		return err
	}
	return storer.Save(ctx, fileName, data)
}

// Read implements services.FileStorage.
func (uc *GenericFileStorage) Read(ctx context.Context, fileName string) ([]byte, error) {
	if strings.TrimSpace(fileName) == "" {
		return nil, errs.BadRequestDirect("fileName es requerido")
	}
	storer, err := uc.newFileStorer(ctx)
	if err != nil {
		return nil, err
	}
	return storer.Read(ctx, fileName)
}

// Delete implements services.FileStorage.
func (uc *GenericFileStorage) Delete(ctx context.Context, fileNames []string) error {
	if len(fileNames) == 0 {
		return nil
	}
	storer, err := uc.newFileStorer(ctx)
	if err != nil {
		return err
	}
	for _, name := range fileNames {
		errors.Join(err, storer.Delete(ctx, name))
	}
	return err
}
