module.exports = {
  "branches": [
    "master",
  ],
  "verifyConditions": [
    "@semantic-release/changelog",
    "@semantic-release/git"
  ],
  "plugins": [
    "@semantic-release/commit-analyzer",
    "@semantic-release/release-notes-generator",
    [
      "@google/semantic-release-replace-plugin",
      {
        "replacements": [{
          "files": ["cmd/root.go"],
          "from": "__VERSION__ = \".*\"",
          "to": "__VERSION__ = \"${nextRelease.version}\"",
          "results": [{
            "file": "cmd/root.go",
            "hasChanged": true,
            "numMatches": 1,
            "numReplacements": 1
          }],
          "countMatches": true
        }]
      }
    ]
  ],
  "prepare": [
    "@semantic-release/changelog",
    {
      "path": "@semantic-release/git",
      "assets": [
        "CHANGELOG.md",
        process.env.README_LOCATION ? process.env.README_LOCATION : 'README.md',
        "yarn.lock",
        "npm-shrinkwrap.json"
      ],
      "message": "chore(release): <%= nextRelease.version %> - <%= new Date().toISOString().slice(0,10).replace(/-/g,'') %> [skip ci]\n\n<%= nextRelease.notes %>"
    }
  ]
}