package di

import (
	"github.com/innobead/kubefire/pkg/bootstrap"
	"github.com/innobead/kubefire/pkg/bootstrap/versionfinder"
	"github.com/innobead/kubefire/pkg/cluster"
	pkgconfig "github.com/innobead/kubefire/pkg/config"
	"github.com/innobead/kubefire/pkg/node"
	"github.com/innobead/kubefire/pkg/output"
	"github.com/sirupsen/logrus"
	"path"
	"reflect"
	"sync"
)

var (
	lock         = &sync.Mutex{}
	initialized  = false
	bootstrapper bootstrap.Bootstrapper
	container    = map[string]interface{}{}
)

type awareInjectTypes struct {
	awareType  reflect.Type
	injectType reflect.Type
}

func DelayInit(force bool) {
	if !force && initialized {
		return
	}

	lock.Lock()
	defer lock.Unlock()

	if !force && initialized {
		return
	}

	if force {
		logrus.Debugln("forcibly initializing dependency injection system")
		initialized = false
		container = map[string]interface{}{}
	} else {
		logrus.Debugln("initializing dependency injection system")
	}

	var awareInjectInterfaceTypes []awareInjectTypes
	var awareInterfaceInstances []interface{}

	// init
	awareInjectInterfaceTypes = append(
		awareInjectInterfaceTypes,
		awareInjectTypes{awareType: reflect.TypeOf(new(ClusterManagerAware)).Elem(), injectType: reflect.TypeOf(new(cluster.Manager)).Elem()},
	)
	awareInjectInterfaceTypes = append(
		awareInjectInterfaceTypes,
		awareInjectTypes{awareType: reflect.TypeOf(new(NodeManagerAware)).Elem(), injectType: reflect.TypeOf(new(node.Manager)).Elem()},
	)
	awareInjectInterfaceTypes = append(
		awareInjectInterfaceTypes,
		awareInjectTypes{awareType: reflect.TypeOf(new(ConfigManagerAware)).Elem(), injectType: reflect.TypeOf(new(pkgconfig.Manager)).Elem()},
	)
	awareInjectInterfaceTypes = append(
		awareInjectInterfaceTypes,
		awareInjectTypes{awareType: reflect.TypeOf(new(BootstrapperAware)).Elem(), injectType: reflect.TypeOf(new(bootstrap.Bootstrapper)).Elem()},
	)
	awareInjectInterfaceTypes = append(
		awareInjectInterfaceTypes,
		awareInjectTypes{awareType: reflect.TypeOf(new(VersionFinderAware)).Elem(), injectType: reflect.TypeOf(new(versionfinder.Finder)).Elem()},
	)
	awareInjectInterfaceTypes = append(
		awareInjectInterfaceTypes,
		awareInjectTypes{awareType: reflect.TypeOf(new(OutputAware)).Elem(), injectType: reflect.TypeOf(new(output.Outputer)).Elem()},
	)

	awareInterfaceInstances = append(awareInterfaceInstances, ClusterManager())
	awareInterfaceInstances = append(awareInterfaceInstances, NodeManager())
	awareInterfaceInstances = append(awareInterfaceInstances, ConfigManager())
	awareInterfaceInstances = append(awareInterfaceInstances, Bootstrapper())
	awareInterfaceInstances = append(awareInterfaceInstances, VersionFinder())
	awareInterfaceInstances = append(awareInterfaceInstances, Output())

	// inject dependencies
	for _, awareInjectInterfaceType := range awareInjectInterfaceTypes {

		for _, awareInterfaceInstance := range awareInterfaceInstances {
			awareInterfaceInstanceValue := reflect.ValueOf(awareInterfaceInstance).Elem().Addr()
			if !awareInterfaceInstanceValue.Type().Implements(awareInjectInterfaceType.awareType) {
				continue
			}

			for i := 0; i < awareInjectInterfaceType.awareType.NumMethod(); i++ {
				m := awareInjectInterfaceType.awareType.Method(i)

				instanceMethod := awareInterfaceInstanceValue.MethodByName(m.Name)
				if instanceMethod.IsValid() {
					key := path.Join(awareInjectInterfaceType.injectType.PkgPath(), awareInjectInterfaceType.injectType.Name())
					injectedObj := reflect.ValueOf(container[key]).Convert(awareInjectInterfaceType.injectType)

					instanceMethod.Call(
						[]reflect.Value{
							injectedObj,
						},
					)
				}
			}
		}
	}

	initialized = true

	logrus.Debugln("completed dependency injection system")
}

func addObjToContainer(key string, obj interface{}) {
	container[key] = obj
}

func getObjFromContainer(key string) interface{} {
	if obj, ok := container[key]; ok {
		return obj
	}

	return nil
}
