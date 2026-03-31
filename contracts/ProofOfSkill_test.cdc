// ProofOfSkill_test.cdc
//
// Tests for the ProofOfSkill contract.
// Run with: flow test contracts/ProofOfSkill_test.cdc

import Test
import "ProofOfSkill"

access(all) let admin = Test.getAccount(0x0000000000000007)

// Deploy the contract before tests.
access(all) fun setup() {
    let err = Test.deployContract(
        name: "ProofOfSkill",
        path: "./ProofOfSkill.cdc",
        arguments: []
    )
    Test.expect(err, Test.beNil())
}

// Test: initial state has zero credentials.
access(all) fun testInitialState() {
    let total = ProofOfSkill.getTotalMinted()
    Test.assertEqual(total, 0 as UInt64)
}

// Test: mint a credential and verify it exists.
access(all) fun testMintCredential() {
    let cid = "QmTestCID123456789012345678901234567890abcdef"
    let score: UInt8 = 85
    let submitter = admin.address

    // Mint via transaction.
    let txResult = Test.Transaction(
        code: Test.readFile("./transactions/mint_credential.cdc"),
        authorizers: [admin.address],
        signers: [admin],
        arguments: [cid, score, submitter]
    )
    Test.expect(txResult, Test.beSucceeded())

    // Verify credential exists.
    let credential = ProofOfSkill.getCredential(cid: cid)
    Test.assert(credential != nil, message: "Credential should exist")
    Test.assertEqual(credential!.cid, cid)
    Test.assertEqual(credential!.score, score)
    Test.assertEqual(credential!.submitter, submitter)
    Test.assertEqual(credential!.id, 1 as UInt64)

    // Verify total minted.
    let total = ProofOfSkill.getTotalMinted()
    Test.assertEqual(total, 1 as UInt64)
}

// Test: hasCredential returns true for existing CID.
access(all) fun testHasCredential() {
    let cid = "QmExistingCID12345678901234567890abcdefghijklm"
    let txResult = Test.Transaction(
        code: Test.readFile("./transactions/mint_credential.cdc"),
        authorizers: [admin.address],
        signers: [admin],
        arguments: [cid, 92 as UInt8, admin.address]
    )
    Test.expect(txResult, Test.beSucceeded())

    Test.assert(
        ProofOfSkill.hasCredential(cid: cid),
        message: "hasCredential should return true"
    )
}

// Test: hasCredential returns false for non-existing CID.
access(all) fun testHasCredentialNotFound() {
    Test.assert(
        !ProofOfSkill.hasCredential(cid: "QmNonExistent"),
        message: "hasCredential should return false for unknown CID"
    )
}

// Test: getCredential returns nil for non-existing CID.
access(all) fun testGetCredentialNotFound() {
    let credential = ProofOfSkill.getCredential(cid: "QmDoesNotExist")
    Test.assert(credential == nil, message: "Should return nil for unknown CID")
}

// Test: getCredentialByID works.
access(all) fun testGetCredentialByID() {
    let cid = "QmByIDTestCID1234567890123456789012345678901234"
    let txResult = Test.Transaction(
        code: Test.readFile("./transactions/mint_credential.cdc"),
        authorizers: [admin.address],
        signers: [admin],
        arguments: [cid, 75 as UInt8, admin.address]
    )
    Test.expect(txResult, Test.beSucceeded())

    let total = ProofOfSkill.getTotalMinted()
    let credential = ProofOfSkill.getCredentialByID(id: total)
    Test.assert(credential != nil, message: "Should find credential by ID")
    Test.assertEqual(credential!.cid, cid)
}

// Test: duplicate CID is rejected.
access(all) fun testDuplicateCIDRejected() {
    let cid = "QmDuplicateTestCID123456789012345678901234567890"

    // First mint should succeed.
    let tx1 = Test.Transaction(
        code: Test.readFile("./transactions/mint_credential.cdc"),
        authorizers: [admin.address],
        signers: [admin],
        arguments: [cid, 80 as UInt8, admin.address]
    )
    Test.expect(tx1, Test.beSucceeded())

    // Second mint with same CID should fail.
    let tx2 = Test.Transaction(
        code: Test.readFile("./transactions/mint_credential.cdc"),
        authorizers: [admin.address],
        signers: [admin],
        arguments: [cid, 90 as UInt8, admin.address]
    )
    Test.expect(tx2, Test.beFailed())
}

// Test: score above 100 is rejected.
access(all) fun testScoreAbove100Rejected() {
    let tx = Test.Transaction(
        code: Test.readFile("./transactions/mint_credential.cdc"),
        authorizers: [admin.address],
        signers: [admin],
        arguments: ["QmScoreTest", 101 as UInt8, admin.address]
    )
    Test.expect(tx, Test.beFailed())
}

// Test: empty CID is rejected.
access(all) fun testEmptyCIDRejected() {
    let tx = Test.Transaction(
        code: Test.readFile("./transactions/mint_credential.cdc"),
        authorizers: [admin.address],
        signers: [admin],
        arguments: ["", 50 as UInt8, admin.address]
    )
    Test.expect(tx, Test.beFailed())
}

// Test: sequential IDs are assigned correctly.
access(all) fun testSequentialIDs() {
    let before = ProofOfSkill.getTotalMinted()

    let tx1 = Test.Transaction(
        code: Test.readFile("./transactions/mint_credential.cdc"),
        authorizers: [admin.address],
        signers: [admin],
        arguments: ["QmSeqTest1_" .concat(before.toString()), 60 as UInt8, admin.address]
    )
    Test.expect(tx1, Test.beSucceeded())

    let tx2 = Test.Transaction(
        code: Test.readFile("./transactions/mint_credential.cdc"),
        authorizers: [admin.address],
        signers: [admin],
        arguments: ["QmSeqTest2_" .concat(before.toString()), 70 as UInt8, admin.address]
    )
    Test.expect(tx2, Test.beSucceeded())

    let after = ProofOfSkill.getTotalMinted()
    Test.assertEqual(after, before + 2)
}
