import { ActorType } from "@app/services/auth/auth-type";
import { MAX_OUTBOX_PAYLOAD_BYTES, TEventInput } from "@app/services/event-outbox/event-outbox-types";

import {
  emitGroupLinkRemoved,
  emitGroupMembershipRemoved,
  emitGroupRemoved,
  emitOrgMembershipRemoved,
  GROUP_LINK_REMOVED_EVENT,
  GROUP_MEMBERSHIP_REMOVED_EVENT,
  GROUP_REMOVED_EVENT,
  GroupMembershipRemovedPayloadSchema,
  GroupRemovedPayloadSchema,
  ORG_MEMBERSHIP_REMOVED_EVENT,
  toOrgMembershipActors
} from "./org-membership-events";

const uuid = (n: number) => `00000000-0000-4000-8000-${String(n).padStart(12, "0")}`;
const TX = { sentinel: true } as never;
const ORG = uuid(1);
const GROUP = uuid(2);
const USER = uuid(10);
const IDENTITY = uuid(20);

const buildEmitter = () => {
  const emitted: { events: TEventInput[]; tx: unknown }[] = [];
  return {
    emitted,
    eventEmitter: {
      emit: async (events: TEventInput[], tx?: unknown) => {
        emitted.push({ events, tx });
      }
    }
  };
};

const expectBatch = (emitted: { events: TEventInput[]; tx: unknown }[], tx: unknown) => {
  expect(emitted).toHaveLength(1);
  expect(emitted[0].tx).toBe(tx);
  return emitted[0].events;
};

describe("org membership events", () => {
  test("toOrgMembershipActors dedupes and tags each kind", () => {
    expect(toOrgMembershipActors({ userIds: [USER, USER], identityIds: [IDENTITY] })).toEqual([
      { actorType: ActorType.USER, actorId: USER },
      { actorType: ActorType.IDENTITY, actorId: IDENTITY }
    ]);
    expect(toOrgMembershipActors({})).toEqual([]);
  });

  test("group membership removed carries the group's orgs and the members who left", async () => {
    const { emitted, eventEmitter } = buildEmitter();
    await emitGroupMembershipRemoved(
      eventEmitter,
      {
        groupId: GROUP,
        orgIds: [ORG, ORG],
        actors: toOrgMembershipActors({ userIds: [USER], identityIds: [IDENTITY] })
      },
      TX
    );

    const events = expectBatch(emitted, TX);
    expect(events).toHaveLength(1);
    expect(events[0].type).toBe(GROUP_MEMBERSHIP_REMOVED_EVENT);
    expect(events[0].id).toMatch(
      /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i
    );
    expect(events[0].data).toEqual({
      groupId: GROUP,
      orgIds: [ORG],
      members: [
        { actorType: ActorType.USER, actorId: USER },
        { actorType: ActorType.IDENTITY, actorId: IDENTITY }
      ]
    });

    const { emitted: none, eventEmitter: noneEmitter } = buildEmitter();
    await emitGroupMembershipRemoved(noneEmitter, { groupId: GROUP, orgIds: [ORG], actors: [] }, TX);
    expect(none).toEqual([]);
  });

  test("group link removed carries its members, paged, and still emits with no members", async () => {
    const { emitted, eventEmitter } = buildEmitter();
    const members = toOrgMembershipActors({ userIds: Array.from({ length: 150 }, (_, i) => uuid(5000 + i)) });
    await emitGroupLinkRemoved(eventEmitter, { orgId: ORG, groupId: GROUP, members }, TX);

    const linkEvents = emitted.flatMap(({ events, tx }) => {
      expect(tx).toBe(TX);
      return events;
    });
    expect(linkEvents).toHaveLength(2);
    for (const event of linkEvents) {
      expect(event.type).toBe(GROUP_LINK_REMOVED_EVENT);
      expect(event.data).toMatchObject({ orgId: ORG, groupId: GROUP });
      expect((event.data as { members: unknown[] }).members.length).toBeLessThanOrEqual(100);
    }

    const { emitted: empty, eventEmitter: emptyEmitter } = buildEmitter();
    await emitGroupLinkRemoved(emptyEmitter, { orgId: ORG, groupId: GROUP, members: [] }, TX);
    const emptyEvents = expectBatch(empty, TX);
    expect(emptyEvents).toHaveLength(1);
    expect(emptyEvents[0].type).toBe(GROUP_LINK_REMOVED_EVENT);
    expect(emptyEvents[0].data).toEqual({ orgId: ORG, groupId: GROUP, members: [] });
  });

  test("group removed still emits for an empty group with only its own org", async () => {
    const { emitted, eventEmitter } = buildEmitter();
    await emitGroupRemoved(eventEmitter, { groupId: GROUP, orgIds: [ORG], members: [] });
    const removedEvents = expectBatch(emitted, undefined);
    expect(removedEvents).toHaveLength(1);
    expect(removedEvents[0].type).toBe(GROUP_REMOVED_EVENT);
    expect(removedEvents[0].data).toEqual({ groupId: GROUP, orgIds: [ORG], members: [] });
  });

  describe.each([
    ["group membership removed", GROUP_MEMBERSHIP_REMOVED_EVENT, GroupMembershipRemovedPayloadSchema],
    ["group removed", GROUP_REMOVED_EVENT, GroupRemovedPayloadSchema]
  ])("%s with a group linked into hundreds of sub-orgs", (_, eventType, schema) => {
    test("pages both lists so every payload fits the outbox and every member is rechecked in every org", async () => {
      const { emitted, eventEmitter } = buildEmitter();
      const orgIds = Array.from({ length: 450 }, (__, i) => uuid(1000 + i));
      const members = toOrgMembershipActors({ userIds: Array.from({ length: 250 }, (__, i) => uuid(5000 + i)) });

      if (eventType === GROUP_REMOVED_EVENT) {
        await emitGroupRemoved(eventEmitter, { groupId: GROUP, orgIds, members }, TX);
      } else {
        await emitGroupMembershipRemoved(eventEmitter, { groupId: GROUP, orgIds, actors: members }, TX);
      }

      const pagedEvents = emitted.flatMap(({ events, tx }) => {
        expect(tx).toBe(TX);
        return events;
      });
      expect(pagedEvents).toHaveLength(5 * 3);
      const covered = new Set<string>();
      for (const event of pagedEvents) {
        expect(event.type).toBe(eventType);
        expect(Buffer.byteLength(JSON.stringify(event.data), "utf8")).toBeLessThan(MAX_OUTBOX_PAYLOAD_BYTES);
        const payload = schema.parse(event.data);
        for (const orgId of payload.orgIds) {
          for (const member of payload.members) covered.add(`${orgId}:${member.actorId}`);
        }
      }
      expect(covered.size).toBe(orgIds.length * members.length);
    });
  });

  test("org membership removed is one event per actor", async () => {
    const { emitted, eventEmitter } = buildEmitter();
    await emitOrgMembershipRemoved(eventEmitter, { orgId: ORG, actors: toOrgMembershipActors({ userIds: [USER] }) });
    const singleEvents = expectBatch(emitted, undefined);
    expect(singleEvents).toHaveLength(1);
    expect(singleEvents[0].type).toBe(ORG_MEMBERSHIP_REMOVED_EVENT);
    expect(singleEvents[0].data).toEqual({ orgId: ORG, actorType: ActorType.USER, actorId: USER });
  });
});
