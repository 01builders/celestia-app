package module

import (
	"context"
	"fmt"

	"cosmossdk.io/core/appmodule"
	pbgrpc "github.com/cosmos/gogoproto/grpc"
	"google.golang.org/grpc"
	protobuf "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	cosmosmsg "cosmossdk.io/api/cosmos/msg/v1"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
)

// Configurator implements the module.Configurator interface.
var _ module.Configurator = Configurator{}

// Configurator is a struct used at startup to register all the message and
// query servers for all modules. It allows the module to register any migrations from
// one consensus version of the module to the next. Finally it maps all the messages
// to the app versions that they are accepted in. This then gets used in the antehandler
// to prevent users from submitting messages that can not yet be executed.
type Configurator struct {
	fromVersion, toVersion uint64
	cdc                    codec.Codec
	err                    error
	msgServer              pbgrpc.Server
	queryServer            pbgrpc.Server
	// acceptedMessages is a map from appVersion -> msgTypeURL -> struct{}.
	acceptedMessages map[uint64]map[string]struct{}
	// migrations is a map of moduleName -> fromVersion -> migration script handler.
	migrations map[string]map[uint64]module.MigrationHandler
}

// NewConfigurator returns a new Configurator instance.
func NewConfigurator(cdc codec.Codec, msgServer, queryServer pbgrpc.Server) Configurator {
	return Configurator{
		cdc:              cdc,
		msgServer:        msgServer,
		queryServer:      queryServer,
		migrations:       map[string]map[uint64]module.MigrationHandler{},
		acceptedMessages: map[uint64]map[string]struct{}{},
	}
}

func (c *Configurator) WithVersions(fromVersion, toVersion uint64) *Configurator {
	c.fromVersion = fromVersion
	c.toVersion = toVersion
	return c
}

// MsgServer implements the Configurator.MsgServer method.
func (c Configurator) MsgServer() pbgrpc.Server {
	return &serverWrapper{
		addMessages: c.addMessages,
		msgServer:   c.msgServer,
	}
}

// GetAcceptedMessages returns the accepted messages for all versions.
// acceptedMessages is a map from appVersion -> msgTypeURL -> struct{}.
func (c Configurator) GetAcceptedMessages() map[uint64]map[string]struct{} {
	return c.acceptedMessages
}

// QueryServer implements the Configurator.QueryServer method.
func (c Configurator) QueryServer() pbgrpc.Server {
	return c.queryServer
}

// RegisterMigration implements the Configurator.RegisterMigration method.
func (c Configurator) RegisterMigration(moduleName string, fromVersion uint64, handler module.MigrationHandler) error {
	if fromVersion == 0 {
		return sdkerrors.ErrInvalidVersion.Wrap("module migration versions should start at 1")
	}

	if c.migrations[moduleName] == nil {
		c.migrations[moduleName] = map[uint64]module.MigrationHandler{}
	}

	if c.migrations[moduleName][fromVersion] != nil {
		return sdkerrors.ErrLogic.Wrapf("another migration for module %s and version %d already exists", moduleName, fromVersion)
	}

	c.migrations[moduleName][fromVersion] = handler

	return nil
}

func (c Configurator) addMessages(msgs []string) {
	for version := c.fromVersion; version <= c.toVersion; version++ {
		if _, exists := c.acceptedMessages[version]; !exists {
			c.acceptedMessages[version] = map[string]struct{}{}
		}
		for _, msg := range msgs {
			c.acceptedMessages[version][msg] = struct{}{}
		}
	}
}

// runModuleMigrations runs all in-place store migrations for one given module from a
// version to another version.
func (c Configurator) runModuleMigrations(ctx sdk.Context, moduleName string, fromVersion, toVersion uint64) error {
	// No-op if toVersion is the initial version or if the version is unchanged.
	if toVersion <= 1 || fromVersion == toVersion {
		return nil
	}

	moduleMigrationsMap, found := c.migrations[moduleName]
	if !found {
		return sdkerrors.ErrNotFound.Wrapf("no migrations found for module %s", moduleName)
	}

	// Run in-place migrations for the module sequentially until toVersion.
	for i := fromVersion; i < toVersion; i++ {
		migrateFn, found := moduleMigrationsMap[i]
		if !found {
			// no migrations needed
			continue
		}
		ctx.Logger().Info(fmt.Sprintf("migrating module %s from version %d to version %d", moduleName, i, i+1))

		err := migrateFn(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c Configurator) Error() error {
	return c.err
}

// Register implements the Configurator.Register method
// It allows to register modules migrations that have migrated to server/v2 but still be compatible with baseapp.
func (c Configurator) Register(moduleName string, fromVersion uint64, handler appmodule.MigrationHandler) error {
	return c.RegisterMigration(moduleName, fromVersion, func(sdkCtx sdk.Context) error {
		return handler(sdkCtx)
	})
}

// RegisterService implements the grpc.Server interface.
func (c Configurator) RegisterService(sd *grpc.ServiceDesc, ss interface{}) {
	c.registerServices(sd, ss)
}

func (c *Configurator) registerServices(sd *grpc.ServiceDesc, ss interface{}) {
	desc, err := c.cdc.InterfaceRegistry().FindDescriptorByName(protoreflect.FullName(sd.ServiceName))
	if err != nil {
		c.err = err
		return
	}

	if protobuf.HasExtension(desc.Options(), cosmosmsg.E_Service) {
		msgs := make([]string, len(sd.Methods))
		for idx, method := range sd.Methods {
			// we execute the handler to extract the message type
			_, _ = method.Handler(nil, context.Background(), func(i interface{}) error {
				msg, ok := i.(sdk.Msg)
				if !ok {
					panic(fmt.Errorf("unable to register service method %s/%s: %T does not implement sdk.Msg", sd.ServiceName, method.MethodName, i))
				}
				msgs[idx] = sdk.MsgTypeURL(msg)
				return nil
			}, noopInterceptor)
		}
		c.addMessages(msgs)
		// call the underlying msg server to actually register the grpc server
		c.msgServer.RegisterService(sd, ss)
	} else {
		c.queryServer.RegisterService(sd, ss)
	}
}
