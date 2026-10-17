# Authentication, registration, and member lifecycle

This document describes the registration and member lifecycle. Email OTP,
member registration/profile completion, MySQL photo storage and retrieval,
Admin/Pengurus role management, and the attestation + Admin verification
workflow are implemented. Phone OTP and contact linking remain unimplemented;
see the backend
[README](../README.md).

## Registration and login

### Email registration

1. A visitor chooses **Daftar** and enters an email address. Email is the only
   OTP channel enabled in this release; phone registration and SMS/WhatsApp OTP
   are not yet available.
2. The visitor requests an email OTP. The address is normalized to lowercase
   and checked for an existing account. The API sends the code through Resend when the
   sender is configured. A verified email already on an account uses this
   same flow for passwordless login; a new address continues registration.
3. The visitor enters the six-digit OTP. The server stores only a keyed hash,
   expires it after 10 minutes, allows at most five guesses, rate-limits sends,
   and marks the contact verified only after a successful check.
4. The visitor completes the required alumni profile: full name, graduation
   year, city/regency, and a profile photo. The city selector uses an
   authoritative list of Indonesian kabupaten/kota rather than free text.
   The registration page loads province and kabupaten/kota choices from
   `api-wilayah-indonesia`; international city names are entered manually.
5. On successful submission, the server creates/activates the member account
   and profile atomically. The initial profile status is `non_verified`; OTP
   verification proves control of a contact, not alumni status.
6. The new member can sign in with any verified contact or linked OAuth
   identity. A verified profile is not required to sign in.
7. Profile completion is shown after registration. A second verified contact
   improves recovery and completeness. Phone OTP is not enabled yet, so this
   step cannot currently be completed in the app. Existing contacts cannot be
   claimed by another account.

The profile completeness indicator is calculated server-side and independent
of alumni-verification status:

| Item | Weight |
| --- | ---: |
| Full name | 20% |
| Graduation year | 20% |
| City/regency | 20% |
| Profile photo | 20% |
| One verified login contact | 10% |
| A second verified contact of the other type | 10% |

The required profile plus one verified login credential yields 90%; a second
verified contact of the other type (for example, phone after email) yields
100%. The UI should identify the missing item, not merely show a percentage.
Since the first release enables email OTP only, the current member flow reports
90% and identifies secondary-contact verification as a later step.

### Google and Facebook

1. The visitor completes the provider's authentication and the site's
   configured anti-abuse challenge.
2. The backend links the provider subject to an account and records provider
   claims only after validating them. A provider identity can authenticate
   the visitor; it does not establish that the visitor is an alumnus.
3. A provider-supplied email is saved as verified only when the provider
   supplies a verified claim. An unverified provider email is saved as
   unverified and confirmed through email OTP before it can be used as a login
   contact.
4. OAuth providers generally do not supply a phone number through ordinary
   login scopes. Do not infer one. Ask the member to enter and OTP-verify a
   number separately.
5. A new OAuth user receives a restricted onboarding session and must complete
   the same mandatory alumni profile; profile submission assigns `member` and
   `non_verified`. Existing provider identities sign in without duplicating
   their account or profile.

### Sign-in and recovery

- A verified email is the normal login identifier; an existing external
  identity is an alternative. Passwordless email login uses the same OTP
  request/verification flow as registration.
- Password login is reserved for explicitly provisioned staging accounts
  unless a separate password-registration and recovery product is approved.
- Login, OTP resend/verification, and recovery endpoints need rate limits,
  non-enumerating responses, same-origin protections, and auditable state
  changes.
- Logout revokes the server-side session. Roles and capabilities are resolved
  from the current account state on every authenticated request.

## Member and staff states

| Account state | How it is obtained | Who may grant/approve it |
| --- | --- | --- |
| Public visitor | No account; can view public content | No account required |
| Onboarding | Authenticated OAuth identity with no alumni profile | Server allows only profile completion |
| Non-verified member | Completes email OTP and required alumni profile, or completes OAuth onboarding | Self-registration; server assigns `member` and `non_verified` |
| Verified member | Completes the alumni verification workflow | At least five attestations from same-batch classmates and final Admin approval |
| Pengurus (officer) | An existing, trusted person is assigned the `officer` role | Admin grants/revokes; the person is verified by Admin |
| Admin | Privileged operational account; never a public signup option | Initial bootstrap is provisioned out of band; subsequent grants require an existing Admin |

The first verified alumni account is a bootstrap exception: an appointed
Pengurus reviews and confirms the initial member to seed the peer-verification
network, with the decision recorded and auditable by Admin. The project owner
has confirmed this one-time exception. It is limited to that bootstrap
account; it is not a public self-selected status or a general bypass of the
five-peer-plus-Admin workflow. Thereafter, the normal verification rule
applies.

There is no fixed Pengurus count. Admin can grant the role to the required
people and revoke it when their duties end. A Pengurus role is separate from
member verification; privileged role grants must be attributed to the Admin
who made them.

## Which accounts must exist first

1. **Initial Admin:** the first Admin identity is restricted to the
   project-owner-supplied name and email in private backend configuration.
   When that mailbox verifies an OTP, the API creates the Admin account once
   and records an audit event. This one-time bootstrap does not create an
   alumni profile and cannot be requested for arbitrary email addresses.
   Public registration never accepts a role. Do not create shared accounts or
   put passwords/provider secrets in source code, a migration, or committed
   configuration. Additional initial Admins must be added by the first Admin
   after the additional people register and verify their email.
2. **Initial Pengurus:** the first Admin verifies the person's identity and
   grants `officer` from the Admin Pengaturan screen after the person has
   registered and verified email. Create as many as operationally needed; no
   fixed number is assumed.
3. **First verified alumni:** the appointed Pengurus reviews the initial
   member; Admin oversees/audits this one-time bootstrap exception. This
   account seeds the initial peer-verification network.
4. **All later alumni:** register themselves and begin as Non-verified.
   Verification follows the ordinary peer-attestation and Admin-approval
   workflow. No alumni account needs to be pre-created.

Staging/demo role accounts are test fixtures only. They must not be copied as
production bootstrap accounts or treated as actual community members.

## Authorization and integrity rules

- The client may request OTP delivery or submit profile data, but cannot
  choose `role`, `verification_status`, `granted_by`, or capabilities.
- A verified contact is unique across accounts. Contact ownership and profile
  changes are checked server-side.
- Peer attestations require a same-batch member with an alumni profile (verified or not), cannot count twice
  for the same target, and must be auditable. The target cannot attest to
  themselves. Admin performs the final decision.
- Officer/Admin access is granted by a server-side role assignment, never by
  OAuth claims, registration input, or frontend state.
- Profile photos are decoded and re-encoded as JPEG/PNG, limited to 2 MiB and
  3,000 pixels per dimension/8 megapixels, then stored in a dedicated MySQL
  table. They are not returned by the directory API; the owner reads theirs at
  `/auth/profile/photo`, and reviewers (Pengurus/Admin, or verified members of
  the same batch) can read the photo of a profile with a *pending* verification
  request.
- The API's existing session cookie and `{message, content}` response
  envelope remain the authentication contract.

## Verification workflow (implemented)

1. A Non-verified member submits a verification request from their profile.
2. Members of the **same batch** (verified or not) see the request and may
   attest once each; self-attestation, duplicates, other batches and accounts
   without an alumni profile are rejected server-side. Only attestations whose
   attester is still active count, and the Admin sees whether each attester is
   verified before deciding.
3. After five valid attestations an Admin approves (or rejects with a note; the
   member may reapply). The Admin cannot approve their own request.
4. Bootstrap exception: while no alumnus is verified, a Pengurus/Admin may
   confirm one pending member without attestations. It is serialized and can be
   used once; the decision is recorded with `decision_kind = bootstrap`.
5. Every step is written to `audit_events`; Admin reads it at
   `GET /api/v1/admin/verifications/audit` (shown under Approval alumni).

## Integration recommendations and decisions

### OTP delivery

Use email as the first OTP channel; the backend currently integrates
[Resend](https://resend.com/pricing), which advertises a free plan
with 3,000 emails/month and a 100-email/day limit; check the provider's current
limits, sender-domain requirements, and terms before launch. This is generally
the simplest low-cost starting point for a small alumni community. Configure
SPF, DKIM, and DMARC for the sending domain and monitor deliverability.

Do not assume WhatsApp OTP is free. The
[WhatsApp Business Platform pricing](https://developers.facebook.com/docs/whatsapp/pricing/)
distinguishes user-initiated service-window messages from business-initiated
template messages. An OTP sent proactively is an authentication template and
can incur a per-message charge; it also requires a WhatsApp Business
integration and approved template. SMS likewise requires a provider and
normally incurs carrier/provider charges. Obtain current Indonesia-specific
quotes before enabling either. For now, email OTP is the default and phone OTP
is unavailable until an SMS or WhatsApp provider is selected.

The sender must be a verified address/domain. Configure `RESEND_API_KEY` and
`RESEND_FROM_EMAIL` only in the private backend environment. `PUBLIC_BASE_URL`
must match the browser origin for same-origin checks. The Resend API key is
also used as the keyed-hash secret for OTPs; rotating it invalidates outstanding
codes. Provider credentials must never be exposed to the browser or logs.

### Profile photos

Photos are files in a persistent directory (`FILE_STORAGE_DIR`; the
`alumni-files` Docker volume in Compose). MySQL keeps only a server-generated
file name (random hex plus `.jpg`/`.png`) and the MIME type in a separate photo
table. The API re-encodes decoded JPEG/PNG, enforces a 2 MiB upload cap and
dimension/pixel limits, writes the file (mode 0600, directory 0700) before the
database transaction, and deletes it if the transaction fails. Photos are never
served as static files: the same authorized endpoints read them from disk, and
file names are validated so a stored value cannot point outside the directory.

On startup with `FILE_STORAGE_DIR` set, photos still stored as blobs are moved
to disk once (idempotent). Without the setting, photos stay in the database as
before.

Back up the photo volume together with MySQL: the database row and the file
must be restored as a pair. Object storage (S3-compatible) remains an option
if the community outgrows a single server.

### Domisili

For Indonesia, the registration UI uses kabupaten/kota IDs from
[api-wilayah-indonesia](https://github.com/emsifa/api-wilayah-indonesia)
instead of free-text input. Its published static v2 endpoint provides
`/provinces.json` and `/regencies/{province_code}.json` without an API key.
The UI validates the chosen regency belongs to the selected province and
persists its stable code with the display name. The API validates the v2
regency-code format and, when `REGENCY_MASTER_PATH` points to a list generated
by `scripts/fetch_regencies.py`, also rejects codes that are not in that master
list; without the file only the format is checked. The browser requires the
external list to load before a domestic city can be selected.

For locations outside Indonesia, provide a country choice and a free-text
city/region field. Store country, location type, code (when domestic), and
display name separately so international text cannot masquerade as an
Indonesian regency code.

### Admin bootstrap and future Admins

Do not use a SQL edit or public sign-up form to make routine Admins. Configure
the first Admin's name/email as `BOOTSTRAP_ADMIN_NAME` and
`BOOTSTRAP_ADMIN_EMAIL` in the private backend environment along with the
Resend sender. The supplied first Admin must verify their own mailbox; no
password is generated or shared.

After bootstrap, an existing Admin should be able to grant/revoke `admin` to
an existing account from the Admin-only Pengaturan screen. The person must
register and verify their email before the Admin can grant access. The API
records `granted_by` and an audit event and prevents removing the last active
Admin. A current authenticated Admin session is required; recent-OTP
reauthentication is a security improvement still to add. Routine role changes
do not require developer or database access.

## Remaining implementation prerequisites

- Configure a Resend API key and verified sender address in the private
  backend environment before email OTP can deliver messages.
- Phone OTP is intentionally unavailable until SMS/WhatsApp delivery is
  selected and configured; secondary-contact completeness therefore remains
  at 90% for now.
- Confirm the school-approved graduation-year range.
- Copy the supplied first Admin name/email to `BOOTSTRAP_ADMIN_NAME` and
  `BOOTSTRAP_ADMIN_EMAIL` in private configuration; no credentials should be
  sent through chat.

Until Resend is configured with a verified sender domain, email OTP endpoints
will return a service-unavailable response. SMS/WhatsApp OTP is not
implemented.
