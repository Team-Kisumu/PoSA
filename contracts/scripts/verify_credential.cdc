// verify_credential.cdc
//
// Script to verify a credential by its CID.
// Read-only — does not require a transaction or signer.

import ProofOfSkill from "../ProofOfSkill.cdc"

access(all) fun main(cid: String): ProofOfSkill.Credential? {
    return ProofOfSkill.getCredential(cid: cid)
}
