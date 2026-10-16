# Staging walkthrough: from Non-verified member to Admin approval

This guide walks through the whole member lifecycle on a staging server using
ready-made accounts. It exists so testers can exercise every role without
needing real email delivery.

## 1. Enable the accounts

In the backend `.env` on the server:

```
STAGING_SCENARIO_PASSWORD=<one password, at least 12 characters>
STAGING_SCENARIO_RESET=true        # only for one start, see "Reset" below
LOGIN_RATE_LIMIT_PER_MINUTE=60     # many testers can share one proxy IP
```

Do **not** configure the four `STAGING_AUTH_*` accounts for this walkthrough:
the seeded "verified" account would count as an existing alumnus and close the
one-time first-member exception (step 2). Keep `DEMO_AUTH_ENABLED=false`.

Rebuild and restart: `docker compose build api && docker compose up -d api`.
All accounts share the same password and sign in on `/login` with
username/password. Share the password out of band; never commit it.

The usernames below are the defaults of the `STAGING_SCENARIO_*_USERNAME` keys
in `.env` (`ADMIN`, `OFFICER_1`, `OFFICER_2`, `MEMBER_1` … `MEMBER_7`); rename
them there if needed. Only `STAGING_SCENARIO_PASSWORD` has to be filled in.

| Username | Role | Notes |
| --- | --- | --- |
| `admin.uji` | Admin | Final approval, role management, audit view |
| `pengurus.satu`, `pengurus.dua` | Pengurus (officer) | Any number of Pengurus can exist; Admin grants the role |
| `alumni.satu` … `alumni.tujuh` | Member, Non-verified | Batch 2012, no photo |

(A real Admin is created separately through `BOOTSTRAP_ADMIN_EMAIL` and email
OTP; `admin.uji` is only a fixture.)

## 2. The flow

1. **Non-verified member asks for verification.** Sign in as `alumni.satu`
   and `alumni.dua`, open **Profil** (`/profil`), and choose **Ajukan
   verifikasi** for each. Status becomes *Menunggu attestasi dan persetujuan*.
2. **First verified alumnus (bootstrap).** Sign in as `pengurus.satu`, open
   **Attestasi alumni**. Because no alumnus is verified yet, **Konfirmasi
   alumni pertama** appears. Confirm `alumni.satu`. They become *Verified*.
   The button disappears for everyone: the exception works once. (`pengurus.dua`
   trying the same on `alumni.dua` gets a 409.)
3. **Five classmates attest `alumni.dua`.** Sign in as `alumni.satu` (now
   Verified), `alumni.tiga`, `alumni.empat`, `alumni.lima` and `alumni.enam`.
   Each opens **Attestasi teman** (`/panel/alumni/attestasi`) and chooses
   **Saya mengenal alumni ini** for `alumni.dua`. Attesters need not be
   verified; they must be in the same batch. Duplicates and self-attestation
   are refused.
4. **Admin decides.** Sign in as `admin.uji`, open **Approval alumni**. The
   request shows 5/5 attestations and who gave them (unverified attesters are
   marked). With fewer than five, **Setujui** stays disabled. Choose
   **Setujui**. `alumni.dua` becomes *Verified*.
5. **Audit.** Below the queue, **Riwayat verifikasi (audit)** lists who
   requested, attested, bootstrapped and approved, with timestamps.
6. **Rejection path (optional).** Have `alumni.tujuh` request verification and
   choose **Tolak** as Admin with a note; the member sees the note and can
   reapply from **Profil**.
7. **Staff management.** As `admin.uji`, open **Pengaturan** to grant or
   revoke Pengurus/Admin for any registered account with a verified email. The
   last active Admin cannot be removed.

## 3. Reset

To rerun from the beginning, set `STAGING_SCENARIO_RESET=true` for one restart:
it deletes verification requests and attestations of the scenario members and
returns them to Non-verified. Set it back to `false` afterwards. Restarting the
API also signs scenario users out. Audit events are kept.

## 4. What this does not cover

Real registration (email OTP, photo, domicile) is exercised separately on
`/registrasi`. Email OTP only reaches arbitrary addresses once a sender domain
is verified in Resend; phone (SMS/WhatsApp) OTP is not implemented.
