#!/bin/bash
# Sync policy_claims from nsi_crawler to nsi_api database
# Usage: docker exec 94-nsip-postgres-1 psql -U postgres -d nsi_api -f /scripts/sync_claims.sql

-- Enable dblink if not exists
CREATE EXTENSION IF NOT EXISTS dblink;

-- Sync claims from crawler DB (only new ones)
INSERT INTO policy_claims (claim_id, policy_id, region_code, policy_type, target_group_tags,
    subsidy_calc_method, subsidy_amount_min, subsidy_amount_max, subsidy_duration,
    effective_date, expire_date, confidence_score, status, version_number,
    conditions, required_documents, source_id, source_name, source_url, policy_url,
    policy_title, issuing_authority, document_number, application_process, contact_info, source_type)
SELECT t.claim_id, t.policy_id, t.region_code, t.policy_type, string_to_array(t.tags, '|'),
    t.subsidy_calc_method, t.min_amt::float8, t.max_amt::float8, t.duration::int,
    t.effective_date::date, NULLIF(t.expire_date,'')::date, t.confidence::float8, t.status, t.ver::int,
    t.conditions::jsonb, t.req_docs::jsonb, t.source_id, t.source_name, t.source_url, t.policy_url,
    t.policy_title, t.issuing_authority, t.document_number, t.app_process::jsonb, t.contact_info, t.source_type
FROM dblink('host=127.0.0.1 dbname=nsi_crawler user=postgres',
    'SELECT claim_id, policy_id, region_code, policy_type, array_to_string(target_group_tags,''|''),
            subsidy_calc_method, subsidy_amount_min::text, subsidy_amount_max::text, subsidy_duration::text,
            effective_date::text, expire_date::text, confidence_score::text, status, version_number::text,
            conditions::text, required_documents::text, source_id, source_name, source_url, policy_url,
            policy_title, issuing_authority, document_number, application_process::text, contact_info, source_type
     FROM policy_claims')
AS t(claim_id text, policy_id text, region_code text, policy_type text, tags text,
     subsidy_calc_method text, min_amt text, max_amt text, duration text,
     effective_date text, expire_date text, confidence text, status text, ver text,
     conditions text, req_docs text, source_id text, source_name text, source_url text, policy_url text,
     policy_title text, issuing_authority text, document_number text, app_process text, contact_info text, source_type text)
WHERE NOT EXISTS (SELECT 1 FROM policy_claims pc WHERE pc.claim_id = t.claim_id);
