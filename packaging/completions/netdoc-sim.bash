# bash completion for netdoc-sim(1)
#
# Hand-maintained: netdoc-sim uses the stdlib flag package, which has no
# completion generator. Keep this in sync with the flag sets in cmd/netdoc-sim.

_netdoc_sim_commands="run lab campaign hunt triage challenge validate scenarios
starters authored capabilities list inspect cleanup version help"

# Bare flag names, per command. Each is spelled both ways below, because the
# flag package accepts -json and --json alike. A command missing from here
# takes no flags at all.
_netdoc_sim_flags() {
    case $1 in
        run) echo "json keep netdoc timeout repeat dry-run v" ;;
        campaign) echo "json runs seed iteration fail-fast netdoc timeout v" ;;
        hunt) echo "json cases seed case shard max-faults generator-version lane fail-fast dry-run netdoc timeout v" ;;
        triage) echo "json scenarios cases hunt-results seed max-faults lane min-severity create context revision netdoc timeout v" ;;
        challenge) echo "id difficulty daily starter authored answer give-up json netdoc timeout v" ;;
        cleanup) echo "all" ;;
        "lab run") echo "all json trace" ;;
    esac
}

_netdoc_sim() {
    local cur prev cmd flag words
    cur=${COMP_WORDS[COMP_CWORD]}
    prev=${COMP_WORDS[COMP_CWORD-1]}

    # netdoc-sim dispatches on argv[1], so the subcommand is always the first
    # word after the binary and nothing before it needs interpreting.
    if ((COMP_CWORD == 1)); then
        COMPREPLY=($(compgen -W "$_netdoc_sim_commands" -- "$cur"))
        return
    fi
    cmd=${COMP_WORDS[1]}
    if [[ $cmd == lab ]]; then
        if ((COMP_CWORD == 2)); then
            COMPREPLY=($(compgen -W "list describe run" -- "$cur"))
        elif [[ ${COMP_WORDS[2]} == run && $cur == -* ]]; then
            words=""
            for flag in $(_netdoc_sim_flags "lab run"); do words+=" -$flag --$flag"; done
            COMPREPLY=($(compgen -W "$words" -- "$cur"))
        elif [[ ${COMP_WORDS[2]} == run || ${COMP_WORDS[2]} == describe ]]; then
            COMPREPLY=($(compgen -W "$(netdoc-sim lab list 2>/dev/null | awk '{print $1}')" -- "$cur"))
        fi
        return
    fi

    case $prev in
        -netdoc | --netdoc)
            COMPREPLY=($(compgen -f -- "$cur"))
            return
            ;;
        -hunt-results | --hunt-results)
            COMPREPLY=($(compgen -d -- "$cur"))
            return
            ;;
        -difficulty | --difficulty)
            COMPREPLY=($(compgen -W "easy medium hard" -- "$cur"))
            return
            ;;
        -min-severity | --min-severity)
            COMPREPLY=($(compgen -W "critical high medium low info" -- "$cur"))
            return
            ;;
        -generator-version | --generator-version)
            COMPREPLY=($(compgen -W "v3 v4 v5 v6 v7" -- "$cur"))
            return
            ;;
        -lane | --lane)
            COMPREPLY=($(compgen -W "bug-oracle stress all" -- "$cur"))
            return
            ;;
        -starter | --starter)
            COMPREPLY=($(compgen -W "fundamentals dns service tls paths routing" -- "$cur"))
            return
            ;;
        -authored | --authored)
            COMPREPLY=($(compgen -W "resolver-refuses resolver-goes-quiet refused-vs-blocked-refused refused-vs-blocked-blocked reset-after-accept certificate-expired certificate-wrong-name no-default-route wrong-default-route missing-subnet-route" -- "$cur"))
            return
            ;;
        -answer | --answer)
            COMPREPLY=($(compgen -W "healthy dns_failure no_default_route wrong_default_route missing_subnet_route preferred_route_failure ipv4_failure ipv6_failure tcp_port_blocked connection_refused connection_reset tls_certificate tls_hostname_mismatch http_service proxy_failure quic_udp_blocked packet_loss" -- "$cur"))
            return
            ;;
        -id | --id | -timeout | --timeout | -repeat | --repeat | -runs | --runs | \
        -seed | --seed | -iteration | --iteration | -cases | --cases | -case | --case | \
        -shard | --shard | \
        -max-faults | --max-faults | -scenarios | --scenarios | -context | --context | \
        -revision | --revision)
            # Ids, durations, numbers and free text: nothing to enumerate, so
            # offer nothing rather than local filenames.
            return
            ;;
    esac

    if [[ $cur == -* ]]; then
        words=""
        for flag in $(_netdoc_sim_flags "$cmd"); do
            words+=" -$flag --$flag"
        done
        COMPREPLY=($(compgen -W "$words" -- "$cur"))
        return
    fi

    # Positional arguments. Simulation ids are not offered: they are printed by
    # the run that kept one, and `netdoc-sim list` is the way to find them.
    case $cmd in
        run | campaign | validate)
            COMPREPLY=($(compgen -W "$(netdoc-sim scenarios 2>/dev/null)" -- "$cur"))
            ;;
        hunt)
            if [[ ${COMP_WORDS[2]} == merge ]]; then
                COMPREPLY=($(compgen -f -- "$cur"))
            else
                COMPREPLY=($(compgen -W "dual-stack-healthy healthy healthy-routed-network socks5h-remote-dns-succeeds tls-valid two-path-healthy two-path-ipv6-healthy two-router-healthy" -- "$cur"))
            fi
            ;;
        starters)
            COMPREPLY=($(compgen -W "fundamentals dns service tls paths routing" -- "$cur"))
            ;;
    esac
}

complete -F _netdoc_sim netdoc-sim
