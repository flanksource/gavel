import { FixtureProfile, FixtureSQLProfile } from '../types';
/** Milliseconds as "125.0 ms" or, from a second, "1.25 s". */
export declare function formatMs(ms: number): string;
/** Bytes with a binary unit, as the fixture benchmark tables report disk and RSS. */
export declare function formatMiB(value: number): string;
export declare function formatSignedBytes(value: number): string;
/** The statements a fixture's SQL profile recorded, behind a "N queries" disclosure. */
export declare function SqlStatements({ profile }: {
    profile: FixtureSQLProfile;
}): import("react").JSX.Element | null;
/** Peak CPU/RSS/memory and observed disk of a fixture step's process tree. */
export declare function ProfileTiles({ profile }: {
    profile: FixtureProfile;
}): import("react").JSX.Element;
//# sourceMappingURL=ResourceParts.d.ts.map