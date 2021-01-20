module.exports = {
  branches: ['master'],
  verifyConditions: ['@semantic-release/changelog', '@semantic-release/git'],
  plugins: [
    '@semantic-release/commit-analyzer',
    '@semantic-release/release-notes-generator',
    [
      '@google/semantic-release-replace-plugin',
      {
        replacements: [
          {
            files: ['version/v.go'],
            from: 'var Version string = ".*"',
            to: 'var Version string = "${nextRelease.version}"',
            results: [
              {
                file: 'version/v.go',
                hasChanged: true,
                numMatches: 1,
                numReplacements: 1
              }
            ],
            countMatches: true
          }
        ]
      }
    ],
    '@semantic-release/changelog',
    [
      '@semantic-release/exec',
      {
        generateNotesCmd: "echo 'latest,${nextRelease.version}' > .tags"
      }
    ],
    [
      '@semantic-release/git',
      {
        assets: ['version/v.go', 'CHANGELOG.md', process.env.README_LOCATION ? process.env.README_LOCATION : 'README.md', 'yarn.lock', 'npm-shrinkwrap.json'],
        message: "chore(release): <%= nextRelease.version %> - <%= new Date().toISOString().slice(0,10).replace(/-/g,'') %> [skip ci]\n\n<%= nextRelease.notes %>"
      }
    ],
    [
      '@semantic-release/github',
      {
        assets: [
          {
            path: 'dist/do-dyndns-linux-amd64',
            label: 'do-dyndns-linux-amd64_v<%= nextRelease.version %>'
          },
          {
            path: 'dist/do-dyndns-linux-armv7',
            label: 'do-dyndns-linux-armv7_v<%= nextRelease.version %>'
          },
          {
            path: 'dist/do-dyndns-linux-arm64',
            label: 'do-dyndns-linux-arm64_v<%= nextRelease.version %>'
          }
        ]
      }
    ]
  ]
}

