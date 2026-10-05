# Docs site

The documentation is plain Markdown in [docs/](docs), rendered with
[md2html](https://github.com/tmc/md2html) and published to GitHub Pages at
https://tmc.github.io/langchaingo/docs/ by
[publish-docs.yaml](../.github/workflows/publish-docs.yaml).

Install md2html and mdvet at the version the workflow uses:

	make install

Preview with live reload at http://localhost:8080:

	make serve

Build the static site into `_site` (pages under `_site/docs`, the files in
[static/](static) at the root):

	make build

Check links, anchors, file name casing, MDX components and included files:

	make vet

Check prose style with Vale:

	make lint-docs

## Authoring

- Navigation order comes from [docs/SUMMARY.md](docs/SUMMARY.md). Add new
  pages there.
- Link to other pages by relative path to the `.md` or `.mdx` file, such
  as `[Agents](../modules/agents/index.mdx)`. md2html rewrites these to
  the rendered pages.
- Pages with `draft: true` in their front matter are not published.
- md2html renders a subset of MDX. Use its components for callouts
  (`<Info>`, `<Note>`, `<Tip>`, `<Warning>`, `<Danger>`), with each tag on
  a line of its own. `make vet` reports components md2html does not know.
- Pages may be `.md` or `.mdx`; both are built the same way. MDX imports,
  exports and JavaScript expressions are not supported, nor is raw HTML.
- To show a program from [examples/](../examples), include it rather than
  copying it, so the page cannot drift from the code:

	<Code src="/examples/llm-chain-example/llm_chain.go" title="examples/llm-chain-example/llm_chain.go" link />

  `src` is relative to the repository root. `link` links the caption to
  the file on GitHub, at the ref named in [docs/docs.json](docs/docs.json).
  `make vet` and `make build` fail if the file does not exist.
