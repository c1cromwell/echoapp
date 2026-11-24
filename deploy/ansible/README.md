# Deploying Echo's chain to real hosts

Ansible for Echo's own network: three Global L0 nodes and three metagraph hosts
(plan: [`docs/OWN_L0_PLAN.md`](../../docs/OWN_L0_PLAN.md) stage 2). It replaces `hydra remote-deploy`,
which only targets Constellation's networks and has no Global L0 playbook.

Every node process is started by one script, [`../sovereign/node/echo-node.sh`](../sovereign/node/echo-node.sh),
as a systemd service `echo-node@<layer>`. The same script runs the local rehearsal
(`make sovereign-up`), so what is tested locally is what runs on the hosts.

## Order of work

1. **Keys** — [`docs/KEY_CEREMONY.md`](../../docs/KEY_CEREMONY.md). Produces the secrets directory and
   `deploy/sovereign/networks/<network>/`.
2. **Hosts** — `deploy/terraform/hetzner` with `-var chain_enabled=true -var metagraph_enabled=true`
   and a `gl0_node_type` (or `-var gl0_colocate=true`). Renders `generated/chain.inventory.yml`.
3. **Binaries** — `metagraph/vendor/fetch.sh`, then in `metagraph/`:
   `sbt currencyL0/assembly currencyL1/assembly dataL1/assembly identityL0/assembly identityL1/assembly`.
4. **Chain** —

   ```bash
   cd deploy/ansible
   ansible-playbook -i ../terraform/hetzner/generated/chain.inventory.yml playbooks/site.yml \
     -e echo_secrets_dir=/Volumes/<drive>/echo-staging
   ```

   The first run performs the three geneses (Global L0, ECHO metagraph, Identity metagraph) and
   writes `metagraph-l0.metagraph-id` and `identity-l0.metagraph-id` into the network directory.
   Commit them. Later runs change nothing on a healthy network.
5. **Gateway** — put the gateway signing key in the Kubernetes secret; the chain addresses are
   already in `deploy/k8s/base/configmap.yaml`.

## Playbooks

| Playbook | Use |
|---|---|
| `site.yml` | Everything, in order. Safe to re-run |
| `global-l0-genesis.yml` | Global L0 only: genesis on first run, otherwise starts stopped nodes |
| `global-l0-join.yml -l gl0-2` | Restart one Global L0 node and wait until it has rejoined. Also the way to roll out a changed seedlist, one node at a time |
| `global-l0-rollback.yml` | Restart the whole Global L0 from its newest snapshot when it can no longer produce snapshots: every node down, or two of three down at once. Finds the newest snapshot itself |
| `metagraph.yml` | Both metagraphs, attached to Echo's Global L0 |
| `status.yml` | State of every layer on every host |
| `verify.yml` | Fails if Global L0 is not including each metagraph's snapshots. Runs at the end of `site.yml`; run it after any restart |

## What is on a host

| Path | Content |
|---|---|
| `/opt/echo` | Launcher, jars, wallet tool |
| `/etc/echo/<layer>.env` | That layer's settings, including the key password. Mode 0640, group `echo` |
| `/etc/echo/keys/` | The host's node key |
| `/etc/echo/*.seedlist`, `gl0.priority` | Who may take part |
| `/var/lib/echo/<layer>/` | Chain data |
| `journalctl -u echo-node@<layer>` | Logs |

## Things to know

- **Seedlists.** `gl0.priority` is the Global L0 nodes; metagraph nodes accept a global snapshot
  only when the majority of them agree. `gl0.seedlist` holds the Global L0 nodes *and* the metagraph
  nodes, because Global L0 accepts a metagraph's snapshot only when a signer is on that list. A
  metagraph node's key could therefore also join Global L0.
- **No chain port is public.** Nodes listen on the private network; the Hetzner firewall admits
  only SSH, and `ufw` on each host admits chain ports only from `10.0.0.0/16`.
- **Environment `dev`.** Deliberate; see the plan §10. The launcher sets it.
- **Signer allowlists.** The launcher refuses to start Metagraph L0 or Identity L0 without the
  gateway's key id, so an open chain cannot be deployed by omission.
- **A metagraph genesis needs a running Global L0.** That is why the genesis happens on the host
  during the deploy and not at the ceremony.

## If `verify.yml` fails: a metagraph cut off from Global L0

Every node can be `Ready` while Global L0 has stopped including a metagraph's snapshots (how this
happens: `docs/OWN_L0_PLAN.md` §12). Nothing anchored meanwhile is final. Restarting one node does
not fix it. This does, for `metagraph-l0` (use `identity-l0` for the Identity metagraph):

```bash
# 1. Stop that layer on every metagraph host.
ansible metagraph -i <inventory> -b -m systemd -a "name=echo-node@metagraph-l0 state=stopped"
# 2. Start it on the first host only. With no peer Ready it resumes from the last snapshot
#    Global L0 accepted. Wait for verify.yml to pass.
ansible 'metagraph[0]' -i <inventory> -b -m systemd -a "name=echo-node@metagraph-l0 state=started"
ansible-playbook -i <inventory> playbooks/verify.yml
# 3. Start the others one at a time, running verify.yml after each.
ansible 'metagraph[1]' -i <inventory> -b -m systemd -a "name=echo-node@metagraph-l0 state=started"
```

Snapshots the metagraph made while cut off are discarded. Anything anchored in that window has to
be submitted again.
