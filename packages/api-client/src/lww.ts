import type { GreetingRecord } from "./types";

export function isNewer(incoming: GreetingRecord, current: GreetingRecord): boolean {
  const incomingAt = Date.parse(incoming.updatedAt);
  const currentAt = Date.parse(current.updatedAt);
  if (incomingAt !== currentAt) {
    return incomingAt > currentAt;
  }
  if (incoming.updatedBy !== current.updatedBy) {
    return incoming.updatedBy > current.updatedBy;
  }
  return incoming.version > current.version;
}
