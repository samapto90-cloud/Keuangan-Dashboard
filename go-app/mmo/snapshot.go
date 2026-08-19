package mmo

func snapshotPayload(world *WorldState) []byte {
	return marshal(TypeWorldSnapshot, world.Snapshot())
}
