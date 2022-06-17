{%hackmd BJrTq20hE %}
# Design DB schema and generate SQL code with dbdiagram.io
###### `simplebank`
Section 1 of [Woring with database](/M50RMDDUS7WN3uDwrZ45lg)

### Design DB schema
1. [introduction to dbdiagram.io](/gWhQXaTjTHCBkgFlF2hMbg)
2. Create DB Schema with [DBML](https://www.dbml.org/docs/#project-definition), a simple, readable DSL language designed to define database strutures.
    1. accounts table
        ```DBML
        Table accounts as A {
          id bigserial [pk]
          owner varchar
          balance bigint
          currency varchar
          created_at timestamptz [default: `now()`]
        }
        ```
        - bigserial: the keyword for auto-increment in Postgres
        - `[pk]` for primary key
        - `timestamptz` for timestamp which includes timezone infomation
        - [default: now()]
    2. entries
        ```DBML
        Table entries {
          id bigserial [pk]
          account_id bigint [ref: > A.id]
          amount bigint
          created_at timestamptz [default: `now()`]
        }
        ```
        - [ref: > A.id] represent many-to-one relationship between entries and accounts table.
        - amount can be positive or negative depending on whether the money is going into or out of the account.
    3. Transfers: records all the money transfers between 2 accounts
        ```DBML
        Table transfers {
          id bigserial [pk]
          from_account_id bigint [ref: > A.id]
          to_account_id bigint [ref: > A.id]
          amount bigint
          created_at timestamptz [default: `now()`]
        }
        ```
        - In the course's case, we just care about internal transfer within simple bank. Thus, using auto-increment id primary key is simple.


    4. add NOT NULL constraint to the tables
    5. add notes to columns with `note`
        ```
        amount bigint  [not null, note: 'can be negative or positive']
        ```
        ![](https://i.imgur.com/mXHuQd6.png)

    6. (Optional) We can use Enum type in DBML, and use it as the type of columns.
        ```
        Enum Currency {
            USD
            EUR
        }
        
        Table accounts as A {
          currency Currency [not null]
          ...
        }
        ```
    7. add indexes to our table by using `index` keyword.
        - We might want to search for accounts by owner name.
            ```
            Indexes {
                owner
            }
            ```
        - Search entries by account_id
            ```
            Indexes {
                account_id
            }
            ```
        - Search transfers by from_account_id, to_account_id
            ```
            Indexes {
                from_account_id
                to_account_id
            }
            ```
        - You can have a composite index
            ```
            Indexes {
                (from_account_id, to_account_id)
            }
            ```
    8. Final table
        ```dbml
        Table accounts as A {
          id bigserial [pk]
          owner varchar [not null]
          balance bigint [not null]
          currency varchar [not null]
          created_at timestamptz [not null, default: `now()`]

          Indexes {
            owner
          }
        }

        Table entries {
          id bigserial [pk]
          account_id bigint [not null, ref: > A.id]
          amount bigint  [not null, note: 'can be negative or positive']
          created_at timestamptz [not null, default: `now()`]
          Indexes {
            account_id
          }
        }

        Table transfers {
          id bigserial [pk]
          from_account_id bigint [ref: > A.id]
          to_account_id bigint [ref: > A.id]
          amount bigint [not null, note: 'must be positive']
          created_at timestamptz [not null, default: `now()`]
          Indexes {
            from_account_id
            to_account_id
            (from_account_id, to_account_id)
          }
        }
        ```


3. Export our DB schema and diagram
    1. click auto-arrange to rearrange our db diagram.
        ![](https://i.imgur.com/41PCboA.png)
        With a small rearrangement.
        ![](https://i.imgur.com/4uVxuat.png)
    2. generate SQL for Postgres by export tools
        ![](https://i.imgur.com/BftzuYN.png)
    3. export DB diagram by export tools
