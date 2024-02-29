mock:
	mockgen -destination pkg/database/mock/users.go github.com/humper/tor_exit_nodes/pkg/database Users
	mockgen -destination pkg/database/mock/tor_exit_nodes.go github.com/humper/tor_exit_nodes/pkg/database TorExitNodes

test:
	go test -coverprofile testcoverage.out -coverpkg ./... ./...

test-coverage: test
	go tool cover -html=testcoverage.out