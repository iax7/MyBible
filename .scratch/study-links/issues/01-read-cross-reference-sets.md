# 01: Read Cross-reference sets

**What to build:** Discover `.crossreferences` Modules and read one Chapter of
their `cross_references` table. A set whose `info.requires_reverse_processing` is
`true` is read in both directions: a link that points at the Chapter comes back
as a link that speaks from it, with its target range and source range swapped.
Book numbers map into the Canonical list, and a link whose target Book is not
Canonical is skipped because it cannot be named. Merge the links of several sets
for one Verse: a target that several links share counts once, the highest-voted
link winning, and the result sorts by votes descending and then by Canonical
target order. Render a link's label as `Book 8:22–24`, with `(from–to)` in front
when the link speaks from a range.

**Blocked by:** nothing.

**Status:** done

- [x] `TypeCrossReferences` discovery lists the installed sets, sorted by abbreviation.
- [x] A set that cannot be read is skipped and reported in the Log.
- [x] `Rows` reads one Chapter from the `cross_references` table.
- [x] `requires_reverse_processing` adds the reversed links with their ranges swapped.
- [x] Legacy book numbers map onto the Canonical list; unknown targets are skipped.
- [x] `CrossReferencesFor` covers source ranges, deduplicates by target and sorts.
- [x] `CrossReference.Label` renders single Verses, ranges and source-range prefixes.
- [x] `modtest.CrossReferences` builds a synthetic set.
- [x] Module tests cover discovery, both directions, ranges, dedupe and labels.
- [x] A real-data check reads `TSK-x` (90 links at Genesis 1:1) and `RST-x` (reversed links and ranges).
