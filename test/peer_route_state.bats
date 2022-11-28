#!/usr/bin/env ./test/libs/bats/bin/bats
load 'common'

@test "verify peer routes announce - embedded" {
  run announce_routes
  sleep 5
  run get_peer_metrics
  # we don't care how many, if one is being withdraw then all should but counter updates take time
  assert_line --regexp '^exabgp_state_route\{.*\} 1$'
}

@test "verify peer routes announce - standalone" {
  run announce_routes
  sleep 5
  run get_peer_metrics 9570
  assert_line --regexp '^exabgp_state_route\{.*\} 1$'
}

@test "verify peer routes ipv4 announce - embedded" {
  run announce_routes
  sleep 5
  run get_peer_metrics
  assert_line --regexp '^exabgp_state_route\{.*family="ipv4 unicast".+nlri="192\.168\.0\.0/24".*\} 1$'
}

@test "verify peer routes ipv6 announce - embedded" {
  run announce_routes
  sleep 5
  run get_peer_metrics
  assert_line --regexp '^exabgp_state_route\{.*family="ipv6 unicast".+nlri="2001:db8:1000::/64".*\} 1$'
}

@test "verify peer routes with as-path, community, local preference and med ipv4 announce - embedded" {
  run announce_routes
  run get_peer_metrics
  assert_line --regexp '^exabgp_state_route\{as_path="65001",communities="65001:1234",family="ipv4 unicast".+,local_preference="100",med="200",nlri="10\.0\.0\.0/24".*\} 1$'
}

@test "verify peer routes with as-path, community, local preference and med ipv6 announce - embedded" {
  run announce_routes
  run get_peer_metrics
  assert_line --regexp '^exabgp_state_route\{as_path="65001",communities="65001:1234",family="ipv6 unicast".+,local_preference="100",med="200",nlri="2001:db8:2000::/64".*\} 1$'
}

@test "verify peer routes with multiple as-paths and communities ipv4 announce - embedded" {
  run announce_routes
  run get_peer_metrics
  assert_line --regexp '^exabgp_state_route\{as_path="65001 65002",communities="65001:1234 65001:5678",family="ipv4 unicast".+,nlri="10\.0\.1\.0/24".*\} 1$'
}

@test "verify peer routes with multiple as-paths and communities ipv6 announce - embedded" {
  run announce_routes
  run get_peer_metrics
  assert_line --regexp '^exabgp_state_route\{as_path="65001 65002",communities="65001:1234 65001:5678",family="ipv6 unicast".+,nlri="2001:db8:3000::/64".*\} 1$'
}

@test "verify peer routes ipv4 announce - standalone" {
  run announce_routes
  sleep 5
  run get_peer_metrics 9570
  assert_line --regexp '^exabgp_state_route\{.*family="ipv4 unicast".+nlri="192\.168\.0\.0/24".*\} 1$'
}

@test "verify peer routes ipv6 announce - standalone" {
  run announce_routes
  sleep 5
  run get_peer_metrics 9570
  assert_line --regexp '^exabgp_state_route\{.*family="ipv6 unicast".+nlri="2001:db8:1000::/64".*\} 1$'
}

@test "verify peer routes with as-path, community, local preference and med ipv4 announce - standalone" {
  run announce_routes
  run get_peer_metrics 9570
  assert_line --regexp '^exabgp_state_route\{as_path="65001",communities="65001:1234",family="ipv4 unicast".+,local_preference="100",med="200",nlri="10\.0\.0\.0/24".*\} 1$'
}

@test "verify peer routes with as-path, community, local preference and med ipv6 announce - standalone" {
  run announce_routes
  run get_peer_metrics 9570
  assert_line --regexp '^exabgp_state_route\{as_path="65001",communities="65001:1234",family="ipv6 unicast".+,local_preference="100",med="200",nlri="2001:db8:2000::/64".*\} 1$'
}

@test "verify peer routes with multiple as-paths and communities ipv4 announce - standalone" {
  run announce_routes
  run get_peer_metrics 9570
  assert_line --regexp '^exabgp_state_route\{as_path="65001 65002",communities="65001:1234 65001:5678",family="ipv4 unicast".+,nlri="10\.0\.1\.0/24".*\} 1$'
}

@test "verify peer routes with multiple as-paths and communities ipv6 announce - standalone" {
  run announce_routes
  run get_peer_metrics 9570
  assert_line --regexp '^exabgp_state_route\{as_path="65001 65002",communities="65001:1234 65001:5678",family="ipv6 unicast".+,nlri="2001:db8:3000::/64".*\} 1$'
}

@test "verify count of peer routes - embedded" {
  run announce_routes
  sleep 5
  run get_route_count
  assert_output '38'
}

@test "verify count of peer routes - standalone" {
  run announce_routes
  sleep 5
  run get_route_count 9570
  assert_output '38'
}

@test "verify peer routes withdraw - embedded" {
  run withdraw_routes
  sleep 5
  run get_peer_metrics
  # we don't care how many, if one is being withdraw then all should but counter updates take time
  assert_line --regexp '^exabgp_state_route\{.*\} 0$'
}

@test "verify peer routes withdraw - standalone" {
  run withdraw_routes
  sleep 5
  run get_peer_metrics 9570
  # standalone exporter should not have any results
  refute_line --regexp '^exabgp_state_route\{.*\} 0$'
}
