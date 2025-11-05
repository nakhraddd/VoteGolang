// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

/**
 * @title VotingLog
 * @dev This contract is owned by the backend application and is used to
 * write immutable, on-chain logs of voting-related activities.
 * It uses events, which are cheap to store and auditable.
 */
contract VotingLog {

    // The address of your backend application's wallet
    address public owner;

    // --- Events ---
    // These events are the on-chain "receipts"

    event CandidateCreated(
        uint indexed candidateId,
        string name,
        string candidateType,
        uint timestamp
    );

    event CandidateVoteCast(
        uint indexed userId,
        uint indexed candidateId,
        string candidateType,
        uint timestamp
    );

    event PetitionCreated(
        uint indexed petitionId,
        uint indexed creatorId,
        string title,
        uint timestamp
    );

    event PetitionVoteCast(
        uint indexed userId,
        uint indexed petitionId,
        string voteType,
        uint timestamp
    );

    /**
     * @dev Sets the 'owner' of the contract to the deployer's address.
     * The 'owner' will be your application's wallet.
     */
    constructor() {
        owner = msg.sender;
    }

    /**
     * @dev Throws if called by any account other than the owner.
     */
    modifier onlyOwner() {
        require(msg.sender == owner, "Only owner can call this function");
        _;
    }

    /**
     * @dev Logs the creation of a new candidate.
     * Only callable by the application wallet.
     */
    function logCandidate(
        uint candidateId,
        string calldata name,
        string calldata candidateType
    ) public onlyOwner {
        emit CandidateCreated(candidateId, name, candidateType, block.timestamp);
    }

    /**
     * @dev Logs a new vote for a candidate.
     * Only callable by the application wallet.
     */
    function logCandidateVote(
        uint userId,
        uint candidateId,
        string calldata candidateType
    ) public onlyOwner {
        emit CandidateVoteCast(userId, candidateId, candidateType, block.timestamp);
    }

    /**
     * @dev Logs the creation of a new petition.
     * Only callable by the application wallet.
     */
    function logPetition(
        uint petitionId,
        uint creatorId,
        string calldata title
    ) public onlyOwner {
        emit PetitionCreated(petitionId, creatorId, title, block.timestamp);
    }

    /**
     * @dev Logs a new vote for a petition.
     * Only callable by the application wallet.
     */
    function logPetitionVote(
        uint userId,
        uint petitionId,
        string calldata voteType
    ) public onlyOwner {
        emit PetitionVoteCast(userId, petitionId, voteType, block.timestamp);
    }
}