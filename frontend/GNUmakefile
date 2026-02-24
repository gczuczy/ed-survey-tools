# ─────────────────────────────────────────────────────────────────────────────
# GNUmakefile – ED Survey Tools
#
# Targets
#   devserver   – install deps (if needed) and start the Angular dev server
#                 with the /api proxy enabled (proxy.conf.json → localhost:8081).
#   build       – production AOT build into dist/
#   clean       – remove node_modules, dist, and Angular caches
# ─────────────────────────────────────────────────────────────────────────────

.PHONY: devserver build clean

# Path to the local npx / ng binary (works even without a global install)
NG   = npx ng
NPM  = npm

# ── devserver ────────────────────────────────────────────────────────────────
# Installs node_modules when they are missing or package.json is newer, then
# fires up `ng serve` which automatically loads proxy.conf.json.
devserver: node_modules
	$(NG) serve --open

node_modules: package.json
	$(NPM) install

# ── build ────────────────────────────────────────────────────────────────────
# Production AOT build.  Output lands in ./dist/ed-survey-tools/browser/
build: node_modules
	$(NG) build --configuration production

# ── clean ────────────────────────────────────────────────────────────────────
clean:
	rm -rf node_modules dist .angular
