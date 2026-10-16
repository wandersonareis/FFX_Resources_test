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

// SchemaBytesLoaders maps version/patternPath keys to avulso loaders: um
// binário APONTADO (bytes crus, fora da árvore do jogo) vira DTO pronto.
// É o decode do acesso externo (menu/CLI arquivo→arquivo).
var SchemaBytesLoaders = map[string]SchemaBytesLoader{}

// SchemaRebuilders maps version/patternPath keys to avulso rebuilders: as
// rows do DTO são aplicadas num binário APONTADO e os bytes reconstruídos
// voltam SEM gravar na árvore — o chamador decide o destino.
var SchemaRebuilders = map[string]SchemaRebuilder{}

// registerSchema registers all maps for the key. A ordem dos argumentos é a
// mesma dos closures no arquivo de schemas: load, loadFrom, apply,
// loadBytes, rebuild, measure.
func registerSchema(key string, load SchemaLoader, loadFrom SchemaSourceLoader, apply SchemaApplier, loadBytes SchemaBytesLoader, rebuild SchemaRebuilder, measure SchemaMeasurer) {
	SchemaRegistry[key] = load
	SchemaSourceLoaders[key] = loadFrom
	SchemaAppliers[key] = apply
	SchemaBytesLoaders[key] = loadBytes
	SchemaRebuilders[key] = rebuild
	SchemaMeasurers[key] = measure
}

// LookupSchema resolve os acessos avulsos por chave de registro: o key é o
// identificador canônico ("ffx/battle/kernel/command.bin"). ok=false =
// binário de objects desconhecido (sem schema registrado).
func LookupSchema(key string) (SchemaBytesLoader, SchemaRebuilder, bool) {
	l, ok := SchemaBytesLoaders[key]
	if !ok {
		return nil, nil, false
	}
	return l, SchemaRebuilders[key], true
}

func init() {
	registerFFXSchemas()
	registerFFX2Schemas()
	registerLastMissSchemas()
}
