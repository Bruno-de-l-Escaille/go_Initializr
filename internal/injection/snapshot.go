package injection

import (
	"go_Initializr/service/snapshot"
	"go_Initializr/service"
)

type snapshotComponents struct {
	snapshotService service.SnapshotServiceInterface
}

func initializeSnapshotComponents(core *coreComponents) *snapshotComponents {
	snapshotService := snapshot.NewSnapshotService(
		core.eventRepository,
		core.eventSubscriber,
		*core.baseRepository,
	)
	return &snapshotComponents{
		snapshotService: snapshotService,
	}
}
