// Static ReDoS screen for caller-supplied regex sources used by the dev-only
// BDD nock harness (bdd-nock-router.dev.ts). Rejects patterns with a quantified
// group that itself contains a quantifier (e.g. (a+)+), the classic
// catastrophic-backtracking shape. The harness only needs simple path
// matchers, so legitimate definitions are unaffected.
// Dependency-free so it can be unit-tested without the server stack.
export const hasNestedQuantifier = (source: string): boolean => {
  const stack: boolean[] = [];
  let inClass = false;
  let escaped = false;
  let innerHasQuantifier = false;
  for (let i = 0; i < source.length; i++) {
    const ch = source[i];
    if (escaped) {
      escaped = false;
      continue;
    }
    if (ch === "\\") {
      escaped = true;
      continue;
    }
    if (ch === "[") {
      inClass = true;
      continue;
    }
    if (ch === "]") {
      inClass = false;
      continue;
    }
    if (inClass) continue;
    if (ch === "(") {
      stack.push(innerHasQuantifier);
      innerHasQuantifier = false;
      continue;
    }
    if (ch === ")") {
      const groupHadQuantifier = innerHasQuantifier;
      // A quantified group nested inside another quantified context is the
      // ReDoS shape even across group boundaries: ((a+))+.
      innerHasQuantifier = (stack.pop() ?? false) || groupHadQuantifier;
      // A quantifier applied to a group that already contains one is only
      // dangerous when the outer repetition is unbounded: (a+)+ but not (a+)?.
      const next = source[i + 1];
      if (groupHadQuantifier && (next === "+" || next === "*")) {
        return true;
      }
      if (groupHadQuantifier && next === "{") {
        const m = /^\{(\d+),(\d*)\}/.exec(source.slice(i + 1));
        // {n} and {n,m} are bounded; {n,} repeats without a cap.
        if (m && m[2] === "") return true;
      }
      // The group's own quantifier counts for an enclosing group.
      if (next === "+" || next === "*" || next === "?" || /^\{\d/.test(source.slice(i + 1))) {
        innerHasQuantifier = true;
      }
      continue;
    }
    if (ch === "+" || ch === "*" || ch === "?") {
      innerHasQuantifier = true;
      continue;
    }
    if (ch === "{") {
      // Only treat {n,m} repetition syntax as a quantifier.
      if (/^\{\d/.test(source.slice(i))) innerHasQuantifier = true;
    }
  }
  return false;
};

// Rejects regex constructs the nested-quantifier screen above does not see
// but that are also classic ReDoS/escape vectors in caller-supplied patterns:
// lookarounds ((?=, (?!, (?<=, (?<!) and backreferences (\1..\9). The BDD
// harness only needs simple path matchers, so legitimate definitions never
// use these. Escaped and character-class occurrences are not constructs.
export const hasRiskyConstruct = (source: string): boolean => {
  let inClass = false;
  let escaped = false;
  for (let i = 0; i < source.length; i++) {
    const ch = source[i];
    if (escaped) {
      // A backreference is an escaped non-zero digit outside a class.
      if (!inClass && ch >= "1" && ch <= "9") return true;
      escaped = false;
      continue;
    }
    if (ch === "\\") {
      escaped = true;
      continue;
    }
    if (ch === "[") {
      inClass = true;
      continue;
    }
    if (ch === "]") {
      inClass = false;
      continue;
    }
    if (inClass) continue;
    if (ch === "(" && source[i + 1] === "?" && (source[i + 2] === "=" || source[i + 2] === "!" || source[i + 2] === "<")) {
      return true;
    }
  }
  return false;
};
