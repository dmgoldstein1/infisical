import { describe, expect, it } from "vitest";

import { hasNestedQuantifier, hasRiskyConstruct } from "./bdd-nock-regex-guard";

describe("bdd-nock regex guard (CodeQL js/regex-injection hardening)", () => {
  it.each([
    "(a+)+",
    "(a*)*",
    "(a|b+)+",
    "((a+))+",
    "(\\w+\\s?)+",
    "(x+){2,}"
  ])("rejects nested-quantifier source %s", (source) => {
    expect(hasNestedQuantifier(source)).toBe(true);
  });

  it.each([
    "^/api/v1/secrets.*$",
    "/define",
    "^/api/v1/[a-z]+/items$",
    "\\d{4}-\\d{2}",
    "(foo|bar)",
    "(ab){2,4}",
    "^/api/(v\\d+/)?secrets([/?].*)?$",
    "/client/v4/zones/[^/]+/dns_records\\?",
    "(a+)?",
    "a+b",
    ""
  ])("allows benign source %s", (source) => {
    expect(hasNestedQuantifier(source)).toBe(false);
  });

  it.each(["(a)(b)\\1", "(?=a)b", "(?!a)b", "(?<=a)b", "(?<!a)b"])("rejects risky construct %s", (source) => {
    expect(hasRiskyConstruct(source)).toBe(true);
  });

  it.each(["^/api/v1/secrets.*$", "(foo|bar)", "\\d+", "[(?=]x", "\\\\1"])("allows benign source for risky screen %s", (source) => {
    expect(hasRiskyConstruct(source)).toBe(false);
  });
});
