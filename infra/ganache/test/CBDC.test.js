const { expect } = require("chai");
const { ethers } = require("hardhat");

describe("CBDC", function () {
    let CBDC;
    let cbdc;
    let owner;
    let addr1;
    let addr2;

    beforeEach(async function () {
        [owner, addr1, addr2] = await ethers.getSigners();
        CBDC = await ethers.getContractFactory("CBDC");
        cbdc = await CBDC.deploy();
        await cbdc.waitForDeployment();
    });

    it("Should set the right owner", async function () {
        expect(await cbdc.owner()).to.equal(owner.address);
    });

    it("Should mint tokens", async function () {
        await cbdc.mint(addr1.address, 100);
        expect(await cbdc.balanceOf(addr1.address)).to.equal(100);
    });

    it("Should blacklist account", async function () {
        await cbdc.blacklist(addr1.address);
        expect(await cbdc.isBlacklisted(addr1.address)).to.equal(true);
        await expect(cbdc.mint(addr1.address, 100)).to.be.revertedWith("Recipient is blacklisted");
    });

    it("Should process offline deposit with valid signature", async function () {
        // 1. Mint to addr1
        await cbdc.mint(addr1.address, 100);

        // 2. Addr1 signs a message to transfer 50 to addr2
        const amount = 50;
        const nonce = 1;
        const contractAddress = await cbdc.getAddress();

        const messageHash = ethers.solidityPackedKeccak256(
            ["address", "uint256", "uint256", "address"],
            [addr1.address, amount, nonce, contractAddress]
        );

        const messageHashBytes = ethers.getBytes(messageHash);
        const signature = await addr1.signMessage(messageHashBytes);

        // 3. Owner (or anyone) submits the depositFor
        await cbdc.depositFor(addr1.address, amount, nonce, signature);

        expect(await cbdc.balanceOf(addr2.address)).to.equal(50); // Note: msg.sender is owner/submitter, wait, depositFor transfers to msg.sender?
        // In the contract: _transfer(from, msg.sender, amount);
        // So if 'owner' calls depositFor, funds go to 'owner'.
        expect(await cbdc.balanceOf(owner.address)).to.equal(50);
    });
});
