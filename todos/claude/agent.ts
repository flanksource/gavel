import { query } from "@anthropic-ai/claude-agent-sdk";
import type { Options, SDKMessage } from "@anthropic-ai/claude-agent-sdk";
import { readFileSync } from "fs";

interface AgentConfig {
  cwd?: string;
  session_id?: string;
  max_budget_usd?: number;
  max_turns?: number;
  model?: string;
  system_prompt?: string;
  tools?: string[];
}

function emit(obj: Record<string, unknown>) {
  process.stdout.write(JSON.stringify(obj) + "\n");
}

async function main() {
  const promptFile = process.argv[2];
  if (!promptFile) {
    emit({ type: "error", message: "Usage: agent.ts <prompt-file>" });
    process.exit(1);
  }

  const prompt = readFileSync(promptFile, "utf-8");
  const config: AgentConfig = JSON.parse(process.env.AGENT_CONFIG || "{}");

  const options: Options = {
    cwd: config.cwd,
    model: config.model,
    maxTurns: config.max_turns || undefined,
    maxBudgetUsd: config.max_budget_usd || undefined,
    permissionMode: "bypassPermissions",
    allowDangerouslySkipPermissions: true,
    allowedTools: config.tools,
    stderr: (data: string) => process.stderr.write(data),
    hooks: {
      PreToolUse: [
        {
          matcher: "Bash",
          hooks: [
            async (input) => {
              const cmd =
                (input as any).tool_input?.command || "";
              if (/\bgit\s+(add|commit)\b/.test(cmd)) {
                return {
                  decision: "block" as const,
                  reason:
                    "git add/commit is managed by gavel, not the agent",
                };
              }
              return { decision: "approve" as const };
            },
          ],
        },
      ],
    },
  };

  if (config.session_id) {
    options.resume = config.session_id;
  }

  if (config.system_prompt) {
    options.systemPrompt = config.system_prompt;
  }

  const stream = query({ prompt, options });

  for await (const message of stream) {
    handleMessage(message);
  }
}

function handleMessage(message: SDKMessage) {
  switch (message.type) {
    case "system":
      if ((message as any).subtype === "init") {
        emit({
          type: "system",
          session_id: message.session_id,
          model: (message as any).model,
          tools: (message as any).tools,
        });
      }
      break;

    case "assistant":
      if ((message as any).message?.content) {
        for (const block of (message as any).message.content) {
          if (block.type === "text") {
            emit({ type: "assistant", text: block.text });
          } else if (block.type === "thinking") {
            emit({ type: "thinking", text: block.thinking });
          } else if (block.type === "tool_use") {
            emit({ type: "tool_use", tool: block.name, input: block.input });
          }
        }
      }
      break;

    case "result":
      emit({
        type: "result",
        success: !(message as any).is_error,
        subtype: (message as any).subtype,
        session_id: message.session_id,
        cost_usd: (message as any).total_cost_usd,
        num_turns: (message as any).num_turns,
        duration_ms: (message as any).duration_ms,
        usage: (message as any).usage,
        errors: (message as any).errors,
        result_text: (message as any).result,
      });
      break;
  }
}

main().catch((err) => {
  emit({ type: "error", message: err?.message || String(err) });
  process.exit(1);
});
