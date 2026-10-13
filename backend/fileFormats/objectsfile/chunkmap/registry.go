package chunkmap

// SchemaRegistry maps version/patternPath keys to chunkmap loaders
// (mods-first). Files present here are served via the chunkmap flow; files
// absent fall through to the legacy KeyedStringFile flow.
var SchemaRegistry = map[string]SchemaLoader{}

// SchemaSourceLoaders maps version/patternPath keys to source-aware loaders
// (used by services.loadOriginalFrom to read the pristine original from
// data/ with the SAME field names as the display collection).
var SchemaSourceLoaders = map[string]SchemaSourceLoader{}

// SchemaAppliers maps version/patternPath keys to chunkmap appliers.
var SchemaAppliers = map[string]SchemaApplier{}

// SchemaMeasurers maps version/patternPath keys to string-table measurers
// (used by services.importObjectsUsage to preview import size limits).
var SchemaMeasurers = map[string]SchemaMeasurer{}

// registerSchema registers all four maps for the key.
func registerSchema(key string, load SchemaLoader, loadFrom SchemaSourceLoader, apply SchemaApplier, measure SchemaMeasurer) {
	SchemaRegistry[key] = load
	SchemaSourceLoaders[key] = loadFrom
	SchemaAppliers[key] = apply
	SchemaMeasurers[key] = measure
}

func init() {
	registerFFXSchemas()
	registerFFX2Schemas()
	registerLastMissSchemas()
}