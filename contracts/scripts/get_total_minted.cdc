// get_total_minted.cdc
//
// Script to get the total number of credentials minted.

import ProofOfSkill from "../ProofOfSkill.cdc"

access(all) fun main(): UInt64 {
    return ProofOfSkill.getTotalMinted()
}
