# AlumniHub ERD

The alumni domain is additive to the existing LMS schema. Existing `accounts` remain the authentication identity; `alumni_profiles.account_id` is a one-to-one foreign key to `accounts.id`.

```mermaid
erDiagram
    ACCOUNTS ||--o| ALUMNI_PROFILES : owns
    ACCOUNTS ||--o{ NEWS_ARTICLES : authors
    ACCOUNTS ||--o{ ALUMNI_EVENTS : creates
    ALUMNI_PROFILES ||--o{ EVENT_REGISTRATIONS : registers
    ALUMNI_EVENTS ||--o{ EVENT_REGISTRATIONS : receives
    ALUMNI_PROFILES ||--o{ BUSINESS_CAREER_LISTINGS : submits
    ACCOUNTS ||--o{ BUSINESS_CAREER_LISTINGS : reviews

    ACCOUNTS {
        uuid id PK
        string email UK
        string name
        int account_type
    }
    ALUMNI_PROFILES {
        uuid id PK
        uuid account_id FK,UK
        int batch_year
        string city
        string occupation
        text bio
        string status
    }
    NEWS_ARTICLES {
        uuid id PK
        uuid author_id FK
        string title
        string slug UK
        string category
        text excerpt
        text body
        string status
        datetime published_at
    }
    ALUMNI_EVENTS {
        uuid id PK
        uuid author_id FK
        string title
        string category
        string venue
        datetime starts_at
        datetime ends_at
        string status
    }
    EVENT_REGISTRATIONS {
        uuid id PK
        uuid event_id FK
        uuid alumni_profile_id FK
        string status
    }
    BUSINESS_CAREER_LISTINGS {
        uuid id PK
        uuid alumni_profile_id FK
        string listing_type
        string name
        string category
        text description
        string contact_url
        string status
        uuid reviewed_by FK
        datetime reviewed_at
    }
    CONTACT_MESSAGES {
        uuid id PK
        string name
        string email
        text message
        string status
    }
```

`status` values: profiles `pending|approved|rejected`; news and events `draft|published|archived`; listings `pending|approved|rejected`; registrations `registered|cancelled|attended`; contact messages `new|read|closed`.