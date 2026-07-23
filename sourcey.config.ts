import { defineConfig, godoc } from "sourcey";

export default defineConfig({
  name: "F-Mesh",
  siteUrl: "https://hovsep.github.io",
  baseUrl: "/fmesh/",
  repo: "https://github.com/hovsep/fmesh",
  editBranch: process.env.SOURCEY_SOURCE_REF ?? "main",
  logo: "./assets/img/logo.png",
  theme: {
    preset: "default",
    colors: {
      primary: "#2563eb",
      light: "#60a5fa",
      dark: "#1d4ed8",
    },
  },
  navigation: {
    tabs: [
      {
        tab: "Go API",
        slug: "api",
        source: godoc({
          module: ".",
          packages: ["./..."],
          mode: "live",
          includeTests: true,
          includeUnexported: false,
          hideUndocumented: false,
        }),
      },
    ],
  },
});
