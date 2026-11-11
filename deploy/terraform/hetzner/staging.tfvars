environment = "staging"
location    = "nbg1"
domain      = "api.staging.echo.app"

# Secrets are NOT stored here. Provide at apply time:
#   export TF_VAR_hcloud_token=...  TF_VAR_ssh_public_key="$(cat ~/.ssh/id_ed25519.pub)"  TF_VAR_db_password=...
# Set the image per release (or via CI):  -var echoapp_image=ghcr.io/c1cromwell/echo-backend:<tag>
