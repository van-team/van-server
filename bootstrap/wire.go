//go:build wireinject
// +build wireinject

package bootstrap

import (
	"github.com/google/wire"
	"github.com/weplanx/server/api"
	"github.com/weplanx/server/common"
)

func NewAPI(values *common.Values) (*api.API, error) {
	wire.Build(
		UseMongoDB,
		UseDatabase,
		UseRedis,
		UseNats,
		UseJetStream,
		UseKeyValue,
		UseSessions,
		UseDSL,
		UsePassport,
		UseLocker,
		UseCaptcha,
		UseHertz,
		UseTransfer,
		api.Provides,
		wire.Struct(new(api.API), "*"),
		wire.Struct(new(common.Inject), "*"),
	)
	return &api.API{}, nil
}
