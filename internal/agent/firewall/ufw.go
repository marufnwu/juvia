package firewall

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

func HandleRuleCreate(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Action   string `json:"action"`
		Port     string `json:"port"`
		Protocol string `json:"protocol"`
		Source   string `json:"source"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if err := validateUFWInstalled(); err != nil {
		return nil, err
	}

	port := strings.TrimSuffix(strings.TrimSuffix(req.Port, "/tcp"), "/udp")

	portWithProto := port
	if req.Protocol != "both" {
		portWithProto = fmt.Sprintf("%s/%s", port, req.Protocol)
	}

	var args []string
	args = append(args, req.Action)

	if req.Source != "any" && req.Source != "0.0.0.0/0" {
		args = append(args, "from", req.Source, "to", "any", "port", portWithProto)
	} else {
		args = append(args, portWithProto)
	}

	cmd := exec.Command("ufw", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("ufw %s failed: %w: %s", req.Action, err, string(output))
	}

	if err := exec.Command("ufw", "--force", "enable").Run(); err != nil {
		return nil, fmt.Errorf("ufw enable failed: %w", err)
	}

	return map[string]interface{}{
		"action":   req.Action,
		"port":     req.Port,
		"protocol": req.Protocol,
		"source":   req.Source,
	}, nil
}

func HandleRuleUpdate(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		OldPort   string `json:"old_port"`
		OldSource string `json:"old_source"`
		Action    string `json:"action"`
		Port      string `json:"port"`
		Protocol  string `json:"protocol"`
		Source    string `json:"source"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if err := validateUFWInstalled(); err != nil {
		return nil, err
	}

	deleteOldRule(req.OldPort, req.OldSource)

	port := strings.TrimSuffix(strings.TrimSuffix(req.Port, "/tcp"), "/udp")

	portWithProto := port
	if req.Protocol != "both" {
		portWithProto = fmt.Sprintf("%s/%s", port, req.Protocol)
	}

	var args []string
	args = append(args, req.Action)

	if req.Source != "any" && req.Source != "0.0.0.0/0" {
		args = append(args, "from", req.Source, "to", "any", "port", portWithProto)
	} else {
		args = append(args, portWithProto)
	}

	cmd := exec.Command("ufw", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("ufw %s failed: %w: %s", req.Action, err, string(output))
	}

	return map[string]interface{}{
		"action":   req.Action,
		"port":     req.Port,
		"protocol": req.Protocol,
		"source":   req.Source,
	}, nil
}

func HandleRuleDelete(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Port   string `json:"port"`
		Source string `json:"source"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if err := validateUFWInstalled(); err != nil {
		return nil, err
	}

	deleteOldRule(req.Port, req.Source)

	return map[string]interface{}{"deleted": true}, nil
}

func deleteOldRule(port, source string) {
	cmd := exec.Command("ufw", "status", "numbered")
	output, _ := cmd.Output()

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if !strings.Contains(line, port) {
			continue
		}
		if !strings.Contains(line, source) {
			continue
		}
		break
	}
}

func validateUFWInstalled() error {
	cmd := exec.Command("which", "ufw")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ufw is not installed")
	}
	return nil
}

func buildUFWArgs(action, port, protocol, source string) []string {
	var args []string
	args = append(args, action)

	if source != "any" {
		args = append(args, "from", source)
	}

	portWithProto := port
	if protocol != "both" {
		portWithProto = fmt.Sprintf("%s/%s", port, protocol)
	}
	args = append(args, "to", portWithProto)

	return args
}