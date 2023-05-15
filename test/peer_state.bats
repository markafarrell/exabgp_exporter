#!/usr/bin/env ./test/libs/bats/bin/bats
load 'common'

@test "verify peer_state is captured - embedded" {
  sleep 5
  run get_peer_metrics
  assert_line --regexp '^exabgp_state_peer\{.*\} [0|1]$'
}

@test "verify peer_state is captured - standalone" {
  sleep 5
  run get_peer_metrics 9570
  assert_line --regexp '^exabgp_state_peer\{.*\} [0|1]$'
}

@test "verify peer_state is down - embedded" {
  run stop_gobgpd
  sleep 2
  run get_peer_metrics
  assert_line --regexp '^exabgp_state_peer\{.*\} 0$'
}

@test "verify peer_state is down - standalone" {
  if [[ $(get_exabgp_version) == "4.2."* ]]; then
    skip "exabgp 4.2.x doesn't report down peers in exabgpcli (issue #996)"
  fi
  run stop_gobgpd
  sleep 2
  run get_peer_metrics 9570
  assert_line --regexp '^exabgp_state_peer\{.*\} 0$'
}

@test "verify peer_state is up - embedded" {
  run start_gobgpd
  sleep 60
  run get_peer_metrics
  assert_line --regexp '^exabgp_state_peer\{.*\} 1$'
  run get_peer_metrics 9570
  assert_line --regexp '^exabgp_state_peer\{.*\} 1$'
}

@test "verify peer_state is up - standalone" {
  run start_gobgpd
  sleep 60
  run get_peer_metrics 9570
  assert_line --regexp '^exabgp_state_peer\{.*\} 1$'
}