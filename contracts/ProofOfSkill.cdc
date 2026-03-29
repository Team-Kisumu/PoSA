// ProofOfSkill.cdc
//
// PoSA Proof-of-Skill credential contract for the Flow blockchain.
// Stores AI evaluation results as tamper-proof, verifiable credentials.
//
// Each credential contains:
//   - cid: IPFS/Filecoin content identifier of the evaluation report
//   - score: AI evaluation score (0-100)
//   - submitter: address of the user who submitted the work
//   - timestamp: block timestamp when the credential was minted
//
// Only the contract admin (deployer) can mint credentials.
// Anyone can verify a credential by its CID.

access(all) contract ProofOfSkill {

    // --- Events ---

    // Emitted when a new credential is minted.
    access(all) event CredentialMinted(
        id: UInt64,
        cid: String,
        score: UInt8,
        submitter: Address,
        timestamp: UFix64
    )

    // --- Credential Struct ---

    // A verifiable proof-of-skill credential.
    access(all) struct Credential {
        access(all) let id: UInt64
        access(all) let cid: String
        access(all) let score: UInt8
        access(all) let submitter: Address
        access(all) let timestamp: UFix64

        init(
            id: UInt64,
            cid: String,
            score: UInt8,
            submitter: Address,
            timestamp: UFix64
        ) {
            self.id = id
            self.cid = cid
            self.score = score
            self.submitter = submitter
            self.timestamp = timestamp
        }
    }

    // --- Storage ---

    // Total number of credentials minted.
    access(all) var totalMinted: UInt64

    // Map of CID -> Credential for lookup.
    access(self) var credentials: {String: Credential}

    // Map of credential ID -> Credential for sequential access.
    access(self) var credentialsByID: {UInt64: Credential}

    // --- Admin Resource ---

    // Admin resource that controls minting. Only the contract deployer
    // receives this resource, restricting minting to the authorized backend.
    access(all) resource Admin {

        // Mint a new proof credential.
        // Requires a valid CID (non-empty) and score (0-100).
        access(all) fun mintCredential(
            cid: String,
            score: UInt8,
            submitter: Address
        ): Credential {
            pre {
                cid.length > 0: "CID must not be empty"
                score <= 100: "Score must be between 0 and 100"
                ProofOfSkill.credentials[cid] == nil: "Credential already exists for this CID"
            }

            let id = ProofOfSkill.totalMinted + 1
            let timestamp = getCurrentBlock().timestamp

            let credential = Credential(
                id: id,
                cid: cid,
                score: score,
                submitter: submitter,
                timestamp: timestamp
            )

            ProofOfSkill.credentials[cid] = credential
            ProofOfSkill.credentialsByID[id] = credential
            ProofOfSkill.totalMinted = id

            emit CredentialMinted(
                id: id,
                cid: cid,
                score: score,
                submitter: submitter,
                timestamp: timestamp
            )

            return credential
        }
    }

    // --- Public Functions ---

    // Look up a credential by its CID.
    // Returns nil if no credential exists for the given CID.
    access(all) fun getCredential(cid: String): Credential? {
        return self.credentials[cid]
    }

    // Look up a credential by its sequential ID.
    access(all) fun getCredentialByID(id: UInt64): Credential? {
        return self.credentialsByID[id]
    }

    // Check if a credential exists for a given CID.
    access(all) fun hasCredential(cid: String): Bool {
        return self.credentials[cid] != nil
    }

    // Get the total number of credentials minted.
    access(all) fun getTotalMinted(): UInt64 {
        return self.totalMinted
    }

    // --- Contract Initialization ---

    init() {
        self.totalMinted = 0
        self.credentials = {}
        self.credentialsByID = {}

        // Store the Admin resource in the deployer's account.
        // Only the deployer can mint credentials.
        let admin <- create Admin()
        self.account.storage.save(<-admin, to: /storage/ProofOfSkillAdmin)
    }
}
