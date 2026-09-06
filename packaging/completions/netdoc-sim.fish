# fish completion for netdoc-sim(1)
#
# Hand-maintained: netdoc-sim uses the stdlib flag package, which has no
# completion generator. Keep this in sync with the flag sets in cmd/netdoc-sim.

# Simulation ids and free text are not enumerable, so suppress the file
# completion fish would otherwise offer; the flags that take a path ask for it
# back with -F.
complete -c netdoc-sim -f

# Subcommands. netdoc-sim dispatches on argv[1], so these are only offered
# while no subcommand has been typed yet.
complete -c netdoc-sim -n __fish_use_subcommand -a run -d 'Build the network, run the tests, print the report'
complete -c netdoc-sim -n __fish_use_subcommand -a campaign -d 'Run a seeded scenario campaign sequentially'
complete -c netdoc-sim -n __fish_use_subcommand -a hunt -d 'Generate deterministic faults and rank likely bugs'
complete -c netdoc-sim -n __fish_use_subcommand -a triage -d 'Hunt the fixed baselines, reproduce findings, file issues'
complete -c netdoc-sim -n __fish_use_subcommand -a challenge -d 'Diagnose a hidden fault yourself, then let netdoc try'
complete -c netdoc-sim -n __fish_use_subcommand -a validate -d 'Parse and check a scenario without building anything'
complete -c netdoc-sim -n __fish_use_subcommand -a lab -d 'Offline scenario lab (list, describe, run)'
complete -c netdoc-sim -n __fish_use_subcommand -a scenarios -d 'List the built-in scenarios'
complete -c netdoc-sim -n __fish_use_subcommand -a starters -d 'List the curated starter packs'
complete -c netdoc-sim -n __fish_use_subcommand -a authored -d 'List the hand-written challenges and their ids'
complete -c netdoc-sim -n __fish_use_subcommand -a capabilities -d 'Report whether this host can simulate'
complete -c netdoc-sim -n __fish_use_subcommand -a list -d 'List simulations left running by run -keep'
complete -c netdoc-sim -n __fish_use_subcommand -a inspect -d 'Show a kept simulation and how to enter it'
complete -c netdoc-sim -n __fish_use_subcommand -a cleanup -d 'Release a kept simulation'
complete -c netdoc-sim -n __fish_use_subcommand -a version -d 'Print the build version'
complete -c netdoc-sim -n __fish_use_subcommand -a help -d 'Print usage and exit'

# Positional arguments. The scenario library is whatever this build knows, so
# ask the binary rather than keeping a copy of the list here.
complete -c netdoc-sim -n 'not __fish_seen_subcommand_from lab; and __fish_seen_subcommand_from run campaign validate' -a '(netdoc-sim scenarios 2>/dev/null)'
complete -c netdoc-sim -n 'not __fish_seen_subcommand_from lab; and __fish_seen_subcommand_from hunt' -a 'dual-stack-healthy healthy healthy-routed-network socks5h-remote-dns-succeeds tls-valid two-path-healthy two-path-ipv6-healthy two-router-healthy'
complete -c netdoc-sim -n 'not __fish_seen_subcommand_from lab; and __fish_seen_subcommand_from starters' -a 'fundamentals dns service tls paths routing'

# The stdlib flag package accepts both spellings of every flag, so each one is
# declared as a short option (-json) and a long option (--json).
complete -c netdoc-sim -n 'not __fish_seen_subcommand_from lab; and __fish_seen_subcommand_from run campaign hunt triage challenge' -o json -l json -d 'Print the machine-readable report'
complete -c netdoc-sim -n 'not __fish_seen_subcommand_from lab; and __fish_seen_subcommand_from run campaign hunt triage challenge' -o netdoc -l netdoc -r -F -d 'The netdoc binary to run'
complete -c netdoc-sim -n 'not __fish_seen_subcommand_from lab; and __fish_seen_subcommand_from run campaign hunt triage challenge' -o timeout -l timeout -r -d 'netdoc per-probe timeout (default 4s)'
complete -c netdoc-sim -n 'not __fish_seen_subcommand_from lab; and __fish_seen_subcommand_from run campaign hunt triage challenge' -o v -l v -d 'Log each privileged command as it runs'
complete -c netdoc-sim -n 'not __fish_seen_subcommand_from lab; and __fish_seen_subcommand_from campaign hunt triage' -o seed -l seed -r -d 'Seed (generated and printed when omitted)'
complete -c netdoc-sim -n 'not __fish_seen_subcommand_from lab; and __fish_seen_subcommand_from hunt triage' -o cases -l cases -r -d 'Unique generated cases to run'
complete -c netdoc-sim -n 'not __fish_seen_subcommand_from lab; and __fish_seen_subcommand_from hunt triage' -o max-faults -l max-faults -r -d 'Maximum mutations per case (default 2, max 3)'
complete -c netdoc-sim -n 'not __fish_seen_subcommand_from lab; and __fish_seen_subcommand_from hunt triage' -o lane -l lane -r -d 'Hunt lane (default bug-oracle)' -a 'bug-oracle stress all'
complete -c netdoc-sim -n 'not __fish_seen_subcommand_from lab; and __fish_seen_subcommand_from campaign hunt' -o fail-fast -l fail-fast -d 'Stop after the first failure'
complete -c netdoc-sim -n 'not __fish_seen_subcommand_from lab; and __fish_seen_subcommand_from run hunt' -o dry-run -l dry-run -d 'Print what would run, without creating namespaces'

complete -c netdoc-sim -n 'not __fish_seen_subcommand_from lab; and __fish_seen_subcommand_from run' -o keep -l keep -d 'Hold the network open after the report'
complete -c netdoc-sim -n 'not __fish_seen_subcommand_from lab; and __fish_seen_subcommand_from run' -o repeat -l repeat -r -d 'Run each test n times (default 1)'

complete -c netdoc-sim -n 'not __fish_seen_subcommand_from lab; and __fish_seen_subcommand_from campaign' -o runs -l runs -r -d 'Override the campaign run count (1-1000)'
complete -c netdoc-sim -n 'not __fish_seen_subcommand_from lab; and __fish_seen_subcommand_from campaign' -o iteration -l iteration -r -d 'Run exactly one derived iteration'

complete -c netdoc-sim -n 'not __fish_seen_subcommand_from lab; and __fish_seen_subcommand_from hunt' -o case -l case -r -d 'Run exactly one derived case'
complete -c netdoc-sim -n 'not __fish_seen_subcommand_from lab; and __fish_seen_subcommand_from hunt' -o shard -l shard -r -d 'Run zero-based shard i/N of the global cases'
complete -c netdoc-sim -n 'not __fish_seen_subcommand_from lab; and __fish_seen_subcommand_from hunt' -o generator-version -l generator-version -r -d 'Hunt generator version (default v7)' -a 'v3 v4 v5 v6 v7'

complete -c netdoc-sim -n 'not __fish_seen_subcommand_from lab; and __fish_seen_subcommand_from triage' -o scenarios -l scenarios -r -d 'Comma-separated baselines (default: all)'
complete -c netdoc-sim -n 'not __fish_seen_subcommand_from lab; and __fish_seen_subcommand_from triage' -o hunt-results -l hunt-results -r -F -d 'Directory of canonical merged hunt reports'
complete -c netdoc-sim -n 'not __fish_seen_subcommand_from lab; and __fish_seen_subcommand_from triage' -o min-severity -l min-severity -r -d 'Lowest severity worth filing (default medium)' -a 'critical high medium low info'
complete -c netdoc-sim -n 'not __fish_seen_subcommand_from lab; and __fish_seen_subcommand_from triage' -o create -l create -d 'File reproducible findings as GitHub issues with gh'
complete -c netdoc-sim -n 'not __fish_seen_subcommand_from lab; and __fish_seen_subcommand_from triage' -o context -l context -r -d 'Debugging context recorded in the issue body'
complete -c netdoc-sim -n 'not __fish_seen_subcommand_from lab; and __fish_seen_subcommand_from triage' -o revision -l revision -r -d 'Commit SHA to record'

complete -c netdoc-sim -n 'not __fish_seen_subcommand_from lab; and __fish_seen_subcommand_from challenge' -o id -l id -r -d 'Replay a specific challenge'
complete -c netdoc-sim -n 'not __fish_seen_subcommand_from lab; and __fish_seen_subcommand_from challenge' -o difficulty -l difficulty -r -d 'Draw a challenge of this difficulty' -a 'easy medium hard'
complete -c netdoc-sim -n 'not __fish_seen_subcommand_from lab; and __fish_seen_subcommand_from challenge' -o daily -l daily -d "Play today's challenge, or -daily=YYYY-MM-DD"
complete -c netdoc-sim -n 'not __fish_seen_subcommand_from lab; and __fish_seen_subcommand_from challenge' -o starter -l starter -r -d 'Draw from a curated starter pack' -a 'fundamentals dns service tls paths routing'
complete -c netdoc-sim -n 'not __fish_seen_subcommand_from lab; and __fish_seen_subcommand_from challenge' -o authored -l authored -r -d 'Play a hand-written challenge by slug' -a 'resolver-refuses resolver-goes-quiet refused-vs-blocked-refused refused-vs-blocked-blocked reset-after-accept certificate-expired certificate-wrong-name no-default-route wrong-default-route missing-subnet-route'
complete -c netdoc-sim -n 'not __fish_seen_subcommand_from lab; and __fish_seen_subcommand_from challenge' -o answer -l answer -r -d 'Submit this diagnosis without opening a shell' -a 'healthy dns_failure no_default_route wrong_default_route missing_subnet_route preferred_route_failure ipv4_failure ipv6_failure tcp_port_blocked connection_refused connection_reset tls_certificate tls_hostname_mismatch http_service proxy_failure quic_udp_blocked packet_loss'
complete -c netdoc-sim -n 'not __fish_seen_subcommand_from lab; and __fish_seen_subcommand_from challenge' -o give-up -l give-up -d 'Skip straight to the answer'

complete -c netdoc-sim -n 'not __fish_seen_subcommand_from lab; and __fish_seen_subcommand_from cleanup' -o all -l all -d 'Release every kept simulation'

# Lab has its own nested grammar, not the namespace run flags.
complete -c netdoc-sim -n '__fish_seen_subcommand_from lab; and not __fish_seen_subcommand_from list describe run' -a 'list describe run'
complete -c netdoc-sim -n '__fish_seen_subcommand_from lab; and __fish_seen_subcommand_from describe run' -a '(netdoc-sim lab list 2>/dev/null | string replace -r " .*" "")'
complete -c netdoc-sim -n '__fish_seen_subcommand_from lab; and __fish_seen_subcommand_from run' -o all -l all -d 'Run every lab scenario'
complete -c netdoc-sim -n '__fish_seen_subcommand_from lab; and __fish_seen_subcommand_from run' -o json -l json -d 'Print experimental lab JSON'
complete -c netdoc-sim -n '__fish_seen_subcommand_from lab; and __fish_seen_subcommand_from run' -o trace -l trace -d 'Print simulator-only packet paths'
