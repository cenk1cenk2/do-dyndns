module.exports = {
  repositoryUrl: 'git@github.com:cenk1cenk2/do-dyndns.git',
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
        // generateNotesCmd: "echo 'latest,${nextRelease.version}' > .tags"
        generateNotesCmd: "echo 'edge' > .tags"
      }
    ],
    [
      '@semantic-release/git',
      {
        assets: ['version/v.go', 'CHANGELOG.md', 'README.md']
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
    ],
    '@semantic-release/gitlab'
  ]
}
