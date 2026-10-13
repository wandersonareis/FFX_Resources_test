import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "export",
  reactCompiler: true,
  experimental: {
    turbopackGc: true,
    turbopackRustReactCompiler: true,
    turbopackLazyDynamicImports: true,
    // Os bindings gerados pelo Wails importam com sufixo ESM
    // (`from "./backend/dto/models.js"`) apontando para arquivos .ts.
    // Sem isso o Turbopack não resolve o specifier e o build quebra.
    extensionAlias: {
      ".js": [".ts", ".tsx", ".js"],
    },
  },
  images: {
    unoptimized: true,
  },
};

export default nextConfig;
