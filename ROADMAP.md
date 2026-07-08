
# Feature Parity with <https://github.com/monobilisim/monokit>

## Database healthchecks

- [ ] __MySql__
  - [X] Up check
  - [X] Process count check
  - [ ] Certification waiting check
  - [ ] Cluster check (available in limeted form on MariaDB)
  - [X] Auto repair with timing
  - [X] PMM check
  - [~] MariaDB support (it is its own module in monokitv2 )

- [ ] __MariaDB__
  - [X] Up check
  - [X] Process count check
  - [ ] Certification waiting check
  - [-] Cluster check
    - [X] Inaccessible clusters check
    - [ ] Cluster status check
    - [ ] Node status check
    - [X] Cluster sync check
    - [ ] Receive queue check
    - [ ] Flow Control check
    - Added Cluster certification check
  - [X] Auto repair with timing
  - [X] PMM check
