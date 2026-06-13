-- Echo Comply portal schema (WO-309).
-- Supabase owns ONLY operator auth + org membership. No compliance data, no message
-- content, no PII beyond an operator email lives here. Compliance data stays in the Go
-- Comply backend / metagraph (system of record).

-- Organizations the portal knows about. `org_did` is the canonical tenant key shared
-- with the Go backend (matches the on-chain org DID).
create table if not exists public.organizations (
  org_did     text primary key,
  name        text not null,
  created_at  timestamptz not null default now()
);

-- Portal roles: who can see/configure what in the console. Distinct from the on-chain
-- EchoOrgRoleCredential (owner/admin/moderator/member), which the backend owns.
create type public.portal_role as enum (
  'owner',
  'admin',
  'compliance_officer',
  'auditor',
  'viewer'
);

-- Operator <-> org membership. `user_id` references the Supabase auth user.
create table if not exists public.org_members (
  user_id    uuid not null references auth.users(id) on delete cascade,
  org_did    text not null references public.organizations(org_did) on delete cascade,
  role       public.portal_role not null default 'viewer',
  created_at timestamptz not null default now(),
  primary key (user_id, org_did)
);

create index if not exists org_members_org_did_idx on public.org_members(org_did);

-- Row-level security: an operator can read only their own membership rows; writes are
-- service-role only (SCIM provisioning / admin console via WO-310).
alter table public.org_members enable row level security;
alter table public.organizations enable row level security;

create policy "members read own memberships"
  on public.org_members for select
  using (auth.uid() = user_id);

create policy "members read their orgs"
  on public.organizations for select
  using (
    exists (
      select 1 from public.org_members m
      where m.org_did = organizations.org_did and m.user_id = auth.uid()
    )
  );
