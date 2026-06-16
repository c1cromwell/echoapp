-- WO-309 SCIM provisioning columns on org_members.

alter table public.org_members
  add column if not exists scim_external_id text,
  add column if not exists active boolean not null default true;

create unique index if not exists org_members_scim_external_id_idx
  on public.org_members(scim_external_id)
  where scim_external_id is not null;
