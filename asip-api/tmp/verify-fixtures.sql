-- Verification fixtures: one fully-mapped prefix (dataset path) and nothing for
-- the Cymru fallback path.
DELETE FROM asn_ipv4;
DELETE FROM country_ipv4;
DELETE FROM asn;
DELETE FROM country;

INSERT INTO country (country, country_code) VALUES
    ('United States', 'US'),
    ('Germany', 'DE');

INSERT INTO asn (asn, "as", country_id)
SELECT 13335, 'Cloudflare Inc.', id FROM country WHERE country_code = 'US';

INSERT INTO ipv4_stat (asn_id, prefixes, prefixes_aggregated, largest_prefix, total_addresses)
SELECT id, 1, 1, 13, 524288 FROM asn WHERE asn = 13335;

INSERT INTO asn_ipv4 (asn_id, cidr, last_modified, start_ip, end_ip)
SELECT id, '104.16.0.0/13', NULL,
       ('104.16.0.0'::inet - '0.0.0.0'::inet)::bigint,
       ('104.23.255.255'::inet - '0.0.0.0'::inet)::bigint
FROM asn WHERE asn = 13335;

INSERT INTO country_ipv4 (country_id, cidr, last_modified, start_ip, end_ip)
SELECT id, '104.16.0.0/13', NULL,
       ('104.16.0.0'::inet - '0.0.0.0'::inet)::bigint,
       ('104.23.255.255'::inet - '0.0.0.0'::inet)::bigint
FROM country WHERE country_code = 'US';

DELETE FROM ip_attribution;

SELECT (SELECT count(*) FROM asn) AS asn_rows,
       (SELECT count(*) FROM asn_ipv4) AS asn_ipv4_rows,
       (SELECT count(*) FROM country_ipv4) AS country_ipv4_rows;
