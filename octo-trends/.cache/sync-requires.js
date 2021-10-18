
// prefer default export if available
const preferDefault = m => (m && m.default) || m


exports.components = {
  "component---cache-dev-404-page-js": preferDefault(require("/home/doda/Dropbox/learnlangs/octogrowth/octo-trends/.cache/dev-404-page.js")),
  "component---src-pages-404-js": preferDefault(require("/home/doda/Dropbox/learnlangs/octogrowth/octo-trends/src/pages/404.js")),
  "component---src-pages-index-js": preferDefault(require("/home/doda/Dropbox/learnlangs/octogrowth/octo-trends/src/pages/index.js")),
  "component---src-pages-make-data-js": preferDefault(require("/home/doda/Dropbox/learnlangs/octogrowth/octo-trends/src/pages/makeData.js"))
}

