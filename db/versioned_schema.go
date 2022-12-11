package db

// VersionedSchema represents versions of the database schema. Each time a change to the schema is introduced,
// a new version should be added, with a key one higher than the last key. The first version should be 1,
// and the version of an uninitialised database is implicitly 0.
type VersionedSchema map[uint16]string
