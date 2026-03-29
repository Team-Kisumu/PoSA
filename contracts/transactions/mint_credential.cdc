// mint_credential.cdc
//
// Transaction to mint a new proof-of-skill credential.
// Must be signed by the account that holds the Admin resource.

import ProofOfSkill from "../ProofOfSkill.cdc"

transaction(cid: String, score: UInt8, submitter: Address) {

    let admin: &ProofOfSkill.Admin

    prepare(signer: auth(BorrowValue) &Account) {
        self.admin = signer.storage.borrow<&ProofOfSkill.Admin>(
            from: /storage/ProofOfSkillAdmin
        ) ?? panic("Admin resource not found — only the contract deployer can mint")
    }

    execute {
        let credential = self.admin.mintCredential(
            cid: cid,
            score: score,
            submitter: submitter
        )
        log("Minted credential #"
            .concat(credential.id.toString())
            .concat(" for CID: ")
            .concat(credential.cid))
    }
}
