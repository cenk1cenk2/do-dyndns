module.exports = {
  branches: ["master"],
  repositoryUrl: "https://github.com/cenk1cenk2/do-dyndns",
  plugins: [
    "@semantic-release/commit-analyzer",
    "@semantic-release/release-notes-generator",
    [
      "@google/semantic-release-replace-plugin",
      {
        replacements: [
          {
            files: ["version/v.go"],
            from: 'var Version string = "(.*)"',
            to: 'var Version string = "v${nextRelease.version}"',
            // results: [
            //   {
            //     file: "version/v.go",
            //     hasChanged: true,
            //     numMatches: 1,
            //     numReplacements: 1,
            //   },
            // ],
            // countMatches: true,
          },
        ],
      },
    ],
    "@semantic-release/changelog",
    [
      "@semantic-release/exec",
      {
        generateNotesCmd: "echo 'latest,v${nextRelease.version}' > .tags",
      },
    ],
    [
      "@semantic-release/git",
      {
        assets: ["version/v.go", "CHANGELOG.md", "README.md"],
      },
    ],
    [
      "@semantic-release/github",
      {
        assets: [
          {
            path: "dist/do-dyndns-linux-amd64",
            label: "do-dyndns-linux-amd64_v<%= nextRelease.version %>",
          },
          {
            path: "dist/do-dyndns-linux-armv7",
            label: "do-dyndns-linux-armv7_v<%= nextRelease.version %>",
          },
          {
            path: "dist/do-dyndns-linux-arm64",
            label: "do-dyndns-linux-arm64_v<%= nextRelease.version %>",
          },
        ],
      },
    ],
  ],
};
