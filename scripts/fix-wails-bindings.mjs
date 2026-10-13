/**
 * Normaliza os imports relativos dos bindings gerados pelo Wails.
 *
 * O `wails3 generate bindings` emite imports no formato ESM/node16
 * (`from "./backend/dto/models.js"`) apontando para arquivos `.ts`. O
 * Turbopack (Next 16) não resolve esse sufixo — `experimental.extensionAlias`
 * existe no Next mas não se aplica ao Turbopack, que só expõe `resolveAlias`
 * e `resolveExtensions` (nenhuma das duas cobre caminhos relativos).
 *
 * Roda logo após a geração, então os bindings continuam intocados na origem e
 * o script é idempotente: sem sufixo `.js`, nada muda.
 *
 * Uso: bun scripts/fix-wails-bindings.mjs
 */
import { readdir, readFile, writeFile } from "node:fs/promises";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const root = resolve(
  dirname(fileURLToPath(import.meta.url)),
  "..",
  "frontend",
  "bindings"
);

// from "./x/y.js" | from '../y.js'  ->  from "./x/y" | from '../y'
const RELATIVE_JS_IMPORT = /(from\s*["'])(\.\.?\/[^"']*?)\.js(["'])/g;

async function* walk(dir) {
  let entries;
  try {
    entries = await readdir(dir, { withFileTypes: true });
  } catch {
    // Diretório ainda não gerado — nada a fazer.
    return;
  }
  for (const entry of entries) {
    const path = join(dir, entry.name);
    if (entry.isDirectory()) {
      yield* walk(path);
    } else if (path.endsWith(".ts") || path.endsWith(".mts")) {
      yield path;
    }
  }
}

let changed = 0;
let total = 0;

for await (const file of walk(root)) {
  total += 1;
  const source = await readFile(file, "utf8");
  const fixed = source.replace(RELATIVE_JS_IMPORT, "$1$2$3");
  if (fixed !== source) {
    await writeFile(file, fixed, "utf8");
    changed += 1;
  }
}

console.log(
  `bindings: ${changed}/${total} arquivo(s) com imports relativos normalizados`
);
