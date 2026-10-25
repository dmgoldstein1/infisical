# Podman-native deployment (quadlet)

Systemd-managed Podman deployment mirroring `docker-compose.prod.yml` —
same images, same hostnames (`db`, `redis` via network aliases), same
published port (80→8080). No compose binary needed: systemd starts the
containers directly from these unit files.

## Install (rootless, per user)

```sh
mkdir -p ~/.config/containers/systemd
cp infisical-*.container infisical-*.network infisical-*.volume ~/.config/containers/systemd/
cp infisical.env.example ~/.config/containers/systemd/infisical.env
$EDITOR ~/.config/containers/systemd/infisical.env   # set a real POSTGRES_PASSWORD
systemctl --user daemon-reload
systemctl --user start infisical-backend.service      # pulls db + redis via Requires=
```

System-wide instead: copy the same files to `/etc/containers/systemd/` (and
`infisical.env` next to them) and use `systemctl` without `--user`.

## Day-2 commands

```sh
systemctl --user status 'infisical-*'        # all three units
podman logs infisical-backend                # app logs
podman auto-update --dry-run                 # what AutoUpdate=registry would pull
systemctl --user stop infisical-backend.service
```

## Notes

- Backend image is deliberately pinned (`v0.167.2`); bump it explicitly after
  checking https://hub.docker.com/r/infisical/infisical/tags — same policy
  as the compose file's PIN comment.
- To run a locally built/patched image: `podman tag` it to the `Image=`
  value or reference its digest; `AutoUpdate=registry` only tracks registry
  tags, never local builds.
- Validate changes with the pipeline's `scripts/verify-podman.sh <tree>`
  (quadlet dry-run + pin/secrets policy checks).
