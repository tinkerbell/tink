on:
  push:
    tags:
      - "v*"
name: Release
env:
  REGISTRY: quay.io
  IMAGE_NAME: ${{ github.repository }}
jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Generate release notes
        run: |
          release_notes=$(gh api repos/{owner}/{repo}/releases/generate-notes -F tag_name=${{ github.ref }} --jq .body)
          echo 'RELEASE_NOTES<<EOF' >> $GITHUB_ENV
          echo "${release_notes}" >> $GITHUB_ENV
          echo 'EOF' >> $GITHUB_ENV
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          OWNER: ${{ github.repository_owner }}
          REPO: ${{ github.event.repository.name }}
      - name: Generate Docker image metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
          flavor: latest=false
          tags: type=ref,event=tag
          github-token: ${{ secrets.GITHUB_TOKEN }}
      - name: Set the FROM_TAG variable
        run: echo "FROM_TAG=sha-${GITHUB_SHA::8}" >> $GITHUB_ENV
      - name: Create tink-server image
        run: skopeo copy --all --dest-creds="${DST_REG_USER}":"${DST_REG_PASS}" docker://"${SRC_IMAGE}" docker://"${DST_IMAGE}"
        env:
          SRC_IMAGE: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:${{ env.FROM_TAG }}
          DST_IMAGE: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:${{ fromJSON(steps.meta.outputs.json).labels['org.opencontainers.image.version'] }}
          DST_REG_USER: ${{ secrets.QUAY_USERNAME }}
          DST_REG_PASS: ${{ secrets.QUAY_PASSWORD }}
      - name: Create tink-controller image
        run: skopeo copy --all --dest-creds="${DST_REG_USER}":"${DST_REG_PASS}" docker://"${SRC_IMAGE}" docker://"${DST_IMAGE}"
        env:
          SRC_IMAGE: ${{ env.REGISTRY }}/tinkerbell/tink-controller:${{ env.FROM_TAG }}
          DST_IMAGE: ${{ env.REGISTRY }}/tinkerbell/tink-controller:${{ fromJSON(steps.meta.outputs.json).labels['org.opencontainers.image.version'] }}
          DST_REG_USER: ${{ secrets.QUAY_USERNAME }}
          DST_REG_PASS: ${{ secrets.QUAY_PASSWORD }}
      - name: Create tink-worker image
        run: skopeo copy --all --dest-creds="${DST_REG_USER}":"${DST_REG_PASS}" docker://"${SRC_IMAGE}" docker://"${DST_IMAGE}"
        env:
          SRC_IMAGE: ${{ env.REGISTRY }}/tinkerbell/tink-worker:${{ env.FROM_TAG }}
          DST_IMAGE: ${{ env.REGISTRY }}/tinkerbell/tink-worker:${{ fromJSON(steps.meta.outputs.json).labels['org.opencontainers.image.version'] }}
          DST_REG_USER: ${{ secrets.QUAY_USERNAME }}
          DST_REG_PASS: ${{ secrets.QUAY_PASSWORD }}
      - name: Create release
        uses: actions/create-release@v1
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        with:
          tag_name: ${{ github.ref }}
          release_name: ${{ github.ref }}
          body: ${{ env.RELEASE_NOTES }}
          draft: false
          prerelease: true
