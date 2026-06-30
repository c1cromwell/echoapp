-- migrations/025_wallet_accounts_pubkey.sql
-- Binds the user-held secp256k1 public key to the DAG wallet address so the
-- backend can verify proof-of-ownership for real-funds custody (the address is
-- derived from this key; the backend never holds the private key).

ALTER TABLE wallet_accounts ADD COLUMN IF NOT EXISTS public_key TEXT;
