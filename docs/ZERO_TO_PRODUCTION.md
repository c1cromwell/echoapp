# Echo — Zero to Production Runbook

**Scope:** the single authoritative path from a clean machine → local test → **staging** → **production**,
for the **messaging backend gateway** (Hetzner k3s), the **Constellation metagraph** (local genesis →
integrationnet → mainnet), and the **iOS app** (TestFlight → App Store). Comply is out of scope here
(see `PRODUCT_LAUNCH.md`).

> ## ⚠️ Read this first
> - **Local steps** are fully runnable today.
> - **Cloud/chain steps** provision real infrastructure and cost money. Every secret/input is a
>   `REPLACE_ME` / terraform variable / k8s Secret — **nothing is hardcoded**. Steps that need live
>   credentials (Hetzner token, DNS, Apple signing, a DAG treasury for staking, seedlist registration,
>   production p12 keys) are marked **⚠️ live**.
> - **Never commit secrets.** Provide them via `TF_VAR_*` env, an out-of-band k8s Secret, or CI secrets.
> - Promote in order: **integrationnet before mainnet**, **staging before prod**. Don't skip rungs.

---

## 0. The environment ladder

| Layer | Local | Staging | Production |
|---|---|---|---|
| Backend gateway | `make dev` (Docker) | Hetzner k3s `staging` | Hetzner k3s `prod` |
| Metagraph | local Hydra `start-genesis` | Constellation **IntegrationNet** | Constellation **MainNet** |
| iOS | Simulator / device | TestFlight | App Store |

Backend infra: `deploy/terraform/hetzner` (dev/staging/prod). App deploy: `deploy/k8s` (kustomize
overlays). Metagraph nodes: `deploy/terraform/hetzner/metagraph-nodes.tf` (opt-in). Image: built/pushed
by `.github/workflows/backend-image.yml` to `ghcr.io/c1cromwell/echo-backend`.

---

## 1. From scratch → local test (do this first)

Follow the existing guides — do **not** duplicate them:
- **Full stack from scratch:** `CONTRIBUTING.md` §1–9 (prereqs → `metagraph/scripts/setup-euclid.sh`
  → `make dev` → build the apps).
- **Metagraph + backend E2E (local genesis):** `docs/metagraph-backend-e2e-testing.md`
  (`scripts/hydra build` → `start-genesis` → `status`, then `make dev`, then `scripts/validate-phase1.sh`).
- **Test gates / sign-off:** `docs/E2E_TESTING.md`.

You are ready to promote when `make dev` is healthy (`make dev-status`), the six metagraph layers answer
on `:9000/:9200/:9300/:9400/:9600/:9500`, and `make validate-phase1` passes.

---

## 2. Backend → staging / production (Hetzner k3s)

### 2.1 Publish the image (⚠️ live: GHCR)
Pushing to `main` or tagging `echo-messaging@vX.Y.Z` triggers `.github/workflows/backend-image.yml`,
publishing `ghcr.io/c1cromwell/echo-backend:<tag>` (`:prod` on release tags). Verify the package appears
under the GitHub org's Packages. To build locally instead: `docker build -f build/Dockerfile -t echo-backend .`.

### 2.2 Provision infra (⚠️ live: Hetzner)
```bash
cd deploy/terraform/hetzner
export TF_VAR_hcloud_token=...                       # Hetzner Cloud API token
export TF_VAR_ssh_public_key="$(cat ~/.ssh/id_ed25519.pub)"
export TF_VAR_db_password='<strong-db-password>'      # reused in the k8s Secret (2.4)
terraform init
terraform apply -var-file=staging.tfvars              # then prod.tfvars, after staging is verified
```
This creates the k3s cluster (traefik disabled, **cert-manager pre-installed**), a Postgres+Redis DB VM
(`10.0.1.50`), a load balancer, and volumes. Grab `terraform output` (`k3s_server_ip`, `load_balancer_ip`).
Fetch kubeconfig from the server: `scp root@<k3s_server_ip>:/etc/rancher/k3s/k3s.yaml ~/.kube/echo-<env>`
and edit its `server:` to the node's public IP.

### 2.3 Install ingress-nginx (⚠️ live, once per cluster)
k3s ships without an ingress controller here (traefik is disabled). Install ingress-nginx (pin a version):
```bash
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.11.2/deploy/static/provider/baremetal/deploy.yaml
```
Point the Hetzner load balancer / DNS at the ingress-nginx NodePort (or use a `LoadBalancer` service).

### 2.4 Apply the app (⚠️ live)
```bash
# 1. Secret (out of band — never committed):
cp deploy/k8s/secret.example.yaml /tmp/echo-secret.yaml && $EDITOR /tmp/echo-secret.yaml
kubectl -n echo-staging apply -f /tmp/echo-secret.yaml && shred -u /tmp/echo-secret.yaml
# 2. ClusterIssuer email — edit deploy/k8s/base/clusterissuer.yaml (REPLACE_ME_EMAIL) once per cluster.
# 3. Deploy (image tag + host + replicas come from the overlay):
kubectl apply -k deploy/k8s/overlays/staging      # or overlays/prod
```
The overlays set the image tag, host (`api.staging.echo.app` / `api.echo.app`), and replica count.
The gateway **auto-runs SQL migrations on boot** (golang-migrate advisory lock → safe across replicas).

### 2.5 DNS + verify (⚠️ live)
Point the `domain` A record at the load balancer IP. Then:
```bash
kubectl -n echo-<env> get pods                       # all Running/Ready
curl -fsS https://api.<env>.echo.app/health          # 200
# WSS handshake:
curl -i -N -H "Connection: Upgrade" -H "Upgrade: websocket" \
  -H "Sec-WebSocket-Version: 13" -H "Sec-WebSocket-Key: dGhlIHNhbXBsZQ==" \
  https://api.<env>.echo.app/ws                       # expect 101 Switching Protocols
```
Certificates are issued automatically by cert-manager via the `letsencrypt-prod` ClusterIssuer.

**Gate:** staging green (health + WSS + a real login/message round-trip) **before** `prod.tfvars`.

---

## 3. iOS → TestFlight → App Store

1. Archive (Release config) — locally or from `release-products.yml` (`echo-messaging@v*` tag). CI produces
   an **unsigned** `.xcarchive`.
2. Signed export (fixes the unsigned-archive gap) — set your Team ID in `ios/ExportOptions.plist`, then:
   ```bash
   xcodebuild -exportArchive -archivePath Echo.xcarchive \
     -exportOptionsPlist ios/ExportOptions.plist -exportPath build/export
   ```
3. Upload the `.ipa` (⚠️ live: Apple) via `xcrun altool --upload-app` (or Xcode Organizer) → it appears in
   App Store Connect (`com.echo.app`).
4. TestFlight bake per `E2E_TESTING.md` §5–§9, then submit for App Store review with a **phased release**.

---

## 4. Metagraph → local genesis → integrationnet → mainnet

> The generic mechanics live in `euclid-development-environment/README.md`
> ("Deploying and Running the Metagraph from Genesis"). This section is the **Echo-specific** procedure.
> Everything runs from inside the Euclid clone (`../euclid-development-environment`).

**Prereqs fixed for you:** `metagraph/euclid.json` now carries the `deploy` (network + gl0 + jvm + ansible)
and `snapshot_fees` blocks and 3 `nodes[]` (they survive `setup-euclid.sh`'s copy). `setup-euclid.sh` also
verifies/restores the `infra/ansible/remote` playbooks that `remote-deploy` needs.

### 4.1 Provision the 3 node VMs + inventory (⚠️ live: Hetzner)
```bash
cd deploy/terraform/hetzner
terraform apply -var-file=staging.tfvars \
  -var metagraph_enabled=true \
  -var 'ansible_ssh_private_key_file=~/.ssh/id_ed25519'
# Renders deploy/terraform/hetzner/generated/hosts.ansible.yml from the live node IPs:
cp generated/hosts.ansible.yml ../../../euclid-development-environment/infra/ansible/remote/hosts.ansible.yml
```
Node size: default `cpx41` (8 vCPU/16GB/240GB) is fine for integrationnet; use a dedicated `ccx33` for
mainnet validators (`-var metagraph_node_type=ccx33`).

### 4.2 Generate production keys (⚠️ live)
```bash
CL_PASSWORD='<strong-pass>' scripts/metagraph/gen-prod-keys.sh
```
Copies 3 p12 keystores into the Euclid clone's `source/p12-files/` and prints each **peer ID**. Update
`metagraph/euclid.json` → `nodes[].key_file.password` to your real password (drop the `"password"`
default), re-run `setup-euclid.sh` to re-copy, and **back up the p12 files offline**.

### 4.3 Select the network + GL0 (edit `metagraph/euclid.json` → `deploy.network`)
- `name`: `integrationnet` (start here) → later `mainnet`.
- `gl0_node.{ip,id,public_port}`: your Global L0 node — Constellation recommends running it with
  **`nodectl`** (not provisioned by terraform). Until filled, the sentinel `:gl0_node_ip` makes
  `create-remote-genesis` refuse to run (by design).

### 4.4 Genesis → deploy → start (⚠️ live)
```bash
cd ../euclid-development-environment
scripts/hydra create-remote-genesis            # → infra/shared/genesis/{genesis.address,genesis.snapshot}
scripts/hydra remote-deploy --force_genesis    # Ansible installs deps + ships JARs to the 3 nodes
scripts/hydra remote-start  --force_genesis    # starts metagraph-l0 / currency-l1 / data-l1
scripts/hydra remote-status
```

### 4.5 IntegrationNet verification, then MainNet gates
- **IntegrationNet:** confirm snapshots at
  `https://be-integrationnet.constellationnetwork.io/currency/<metagraph-id>/snapshots/latest`, and
  `/node/info` on each node. Bake here before mainnet.
- **⚠️ MainNet only (hard gates — do NOT skip):**
  1. **Seedlist:** register every node peer ID (from 4.2) on the metagraph seedlist, or startup is
     rejected.
  2. **Staking:** the token treasury must satisfy the DAG staking requirement (per Constellation; the
     Echo launch plan cites 750K DAG across 3 L0 hybrid nodes — `docs/Echo_Combined_Requirements.md`).
  3. Set `deploy.network.name = mainnet`, re-run 4.4, verify on `be-mainnet.constellationnetwork.io`.

---

## 5. Wire the backend → the production metagraph

The gateway reads `CURRENCY_L1_URL`, `DATA_L1_URL`, `IDENTITY_L0_URL`, `IDENTITY_L1_URL` (see
`pkg/config`). Locally these are localhost; in `deploy/k8s/base/configmap.yaml` they are `REPLACE_ME`.
After 4.4, set them (per overlay) to the production node endpoints, plus the identity service credentials
in the Secret (`IDENTITY_SERVICE_DID`, `IDENTITY_SERVICE_KEY_PEM`), then `kubectl apply -k` and
`kubectl rollout restart deploy/echo-backend`.

---

## 6. Rollback + monitoring

- **Backend:** `kubectl -n echo-<env> rollout undo deploy/echo-backend` (or re-apply the prior image tag).
- **Metagraph:** `scripts/hydra remote-start` supports a rollback/resume path; keep `genesis.snapshot`
  backed up. Run the Euclid monitoring service (`hydra install-monitoring-service`) against the
  `monitoring` host in the inventory.
- **Health/alerting:** alert on `/health` (backend) and `/node/info` + block-explorer snapshot lag
  (metagraph). This satisfies the `GO_LIVE_SEPT_1_2026.md` §3.3 backend prod checklist.

---

## Appendix — what this runbook fixed (was missing before)
- `metagraph/euclid.json` had no `deploy`/`nodes`/`snapshot_fees` → remote deploy config was lost on
  `setup-euclid.sh` copy. **Now present.**
- No backend image was published anywhere. **`.github/workflows/backend-image.yml` added.**
- Hetzner terraform built a cluster but nothing deployed the app. **`deploy/k8s` (kustomize) added.**
- The 3 metagraph node VMs were provisioned nowhere. **`metagraph-nodes.tf` + auto-rendered Ansible
  inventory added.**
- Release CI produced unsigned iOS archives. **`ios/ExportOptions.plist` + export step documented.**
