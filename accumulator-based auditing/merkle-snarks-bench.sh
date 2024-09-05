#! /bin/bash

time cargo build

export BELLMAN_VERBOSE=1
(time cargo +nightly run --release --example set_proof merkle 04 30 --hash pedersen -v --iparams pedersen-30/pedersen-TX-02.in;
time cargo +nightly run --release --example set_proof merkle 08 30 --hash pedersen -v --iparams pedersen-30/pedersen-TX-03.in;
time cargo +nightly run --release --example set_proof merkle 16 30 --hash pedersen -v --iparams pedersen-30/pedersen-TX-04.in;
time cargo +nightly run --release --example set_proof merkle 32 30 --hash pedersen -v --iparams pedersen-30/pedersen-TX-05.in;
time cargo +nightly run --release --example set_proof merkle 64 30 --hash pedersen -v --iparams pedersen-30/pedersen-TX-06.in;
time cargo +nightly run --release --example set_proof merkle 128 30 --hash pedersen -v --iparams pedersen-30/pedersen-TX-07.in;
time cargo +nightly run --release --example set_proof merkle 04 30 --hash poseidon -v --iparams poseidon-30/poseidon-TX-02.in;
time cargo +nightly run --release --example set_proof merkle 08 30 --hash poseidon -v --iparams poseidon-30/poseidon-TX-03.in;
time cargo +nightly run --release --example set_proof merkle 16 30 --hash poseidon -v --iparams poseidon-30/poseidon-TX-04.in;
time cargo +nightly run --release --example set_proof merkle 32 30 --hash poseidon -v --iparams poseidon-30/poseidon-TX-05.in;
time cargo +nightly run --release --example set_proof merkle 64 30 --hash poseidon -v --iparams poseidon-30/poseidon-TX-06.in;
time cargo +nightly run --release --example set_proof merkle 128 30 --hash poseidon -v --iparams poseidon-30/poseidon-TX-07.in;
time cargo +nightly run --release --example set_proof merkle 256 30 --hash poseidon -v --iparams poseidon-30/poseidon-TX-08.in;
time cargo +nightly run --release --example set_proof merkle 512 30 --hash poseidon -v --iparams poseidon-30/poseidon-TX-09.in;
time cargo +nightly run --release --example set_proof merkle 1024 30 --hash poseidon -v --iparams poseidon-30/poseidon-TX-10.in;) | tee merkle-snarks-baseline.log

time python parse-logs.py merkle-snarks-baseline.log
