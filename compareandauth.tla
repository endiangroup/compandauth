--------------------------- MODULE compareandauth ---------------------------
EXTENDS Integers, TLC
CONSTANTS DELTA, MAX_SESSIONS

(* --algorithm compareandauth
variables master_caa = 0, issued_sessions = {}

define
isPositive(x) == x >= 0
abs(x) == IF isPositive(x)
          THEN x
          ELSE -1 * x
step(x, y) == IF isPositive(x)
              THEN x + y
              ELSE x - y
step1(x) == step(x, 1)
max(j, k) == IF j >= k
             THEN j
             ELSE k
min(j, k) == IF j <= k
             THEN j
             ELSE k
isLocked(x) == x < 0
hasIssued(x) == x /= 0
isValid(x) == /\ ~isLocked(master_caa)
              /\ hasIssued(master_caa)
              /\ x + DELTA >= abs(master_caa) 
              /\ x < abs(master_caa)

MasterCaaValueNeverIssued == master_caa \notin issued_sessions
IssuedSessionsAlwaysPositiveIntegers == \A x \in issued_sessions: isPositive(x)
LastDeltaIssuedSessionsAreValid == hasIssued(master_caa) => {x \in issued_sessions: isValid(x)} \subseteq max(0, abs(master_caa)-DELTA) .. max(0, abs(master_caa)-1)
AllOtherIssuedSessionsAreInvalid == master_caa > DELTA => {x \in issued_sessions: ~isValid(x)} \subseteq 0 .. abs(master_caa)-DELTA-1
AllIssuedSessionsLessThanMasterCaa == hasIssued(master_caa) => issued_sessions \subseteq 0 .. abs(master_caa)-1
WhenLockedAllIssuedSessionsAreInvalid == hasIssued(master_caa) /\ isLocked(master_caa) => \A x \in issued_sessions: ~isValid(x)
AnySessionGreaterThanMasterCaaIsInvalid == \A x \in abs(master_caa) .. MAX_SESSIONS: ~isValid(x)
end define;

begin
Act:
    while abs(master_caa) < MAX_SESSIONS do
        either
            Issue: issued_sessions := issued_sessions \union {abs(master_caa)};
                   master_caa := step1(master_caa);
        or
            await ~isLocked(master_caa);
            Lock: master_caa := -1 * abs(master_caa);
        or
            await isLocked(master_caa);
            Unlock: master_caa := abs(master_caa);
        or
            await hasIssued(master_caa);
            Revoke: master_caa := step(master_caa, min(abs(master_caa), DELTA));
        end either; 
    end while;
end algorithm; *)

\* BEGIN TRANSLATION
VARIABLES master_caa, issued_sessions, pc

(* define statement *)
isPositive(x) == x >= 0
abs(x) == IF isPositive(x)
          THEN x
          ELSE -1 * x
step(x, y) == IF isPositive(x)
              THEN x + y
              ELSE x - y
step1(x) == step(x, 1)
max(j, k) == IF j >= k
             THEN j
             ELSE k
min(j, k) == IF j <= k
             THEN j
             ELSE k
isLocked(x) == x < 0
hasIssued(x) == x /= 0
isValid(x) == /\ ~isLocked(master_caa)
              /\ hasIssued(master_caa)
              /\ x + DELTA >= abs(master_caa)
              /\ x < abs(master_caa)

MasterCaaValueNeverIssued == master_caa \notin issued_sessions
IssuedSessionsAlwaysPositiveIntegers == \A x \in issued_sessions: isPositive(x)
LastDeltaIssuedSessionsAreValid == hasIssued(master_caa) => {x \in issued_sessions: isValid(x)} \subseteq max(0, abs(master_caa)-DELTA) .. max(0, abs(master_caa)-1)
AllOtherIssuedSessionsAreInvalid == master_caa > DELTA => {x \in issued_sessions: ~isValid(x)} \subseteq 0 .. abs(master_caa)-DELTA-1
AllIssuedSessionsLessThanMasterCaa == hasIssued(master_caa) => issued_sessions \subseteq 0 .. abs(master_caa)-1
WhenLockedAllIssuedSessionsAreInvalid == hasIssued(master_caa) /\ isLocked(master_caa) => \A x \in issued_sessions: ~isValid(x)
AnySessionGreaterThanMasterCaaIsInvalid == \A x \in abs(master_caa) .. MAX_SESSIONS: ~isValid(x)


vars == << master_caa, issued_sessions, pc >>

Init == (* Global variables *)
        /\ master_caa = 0
        /\ issued_sessions = {}
        /\ pc = "Act"

Act == /\ pc = "Act"
       /\ IF abs(master_caa) < MAX_SESSIONS
             THEN /\ \/ /\ pc' = "Issue"
                     \/ /\ ~isLocked(master_caa)
                        /\ pc' = "Lock"
                     \/ /\ isLocked(master_caa)
                        /\ pc' = "Unlock"
                     \/ /\ hasIssued(master_caa)
                        /\ pc' = "Revoke"
             ELSE /\ pc' = "Done"
       /\ UNCHANGED << master_caa, issued_sessions >>

Issue == /\ pc = "Issue"
         /\ issued_sessions' = (issued_sessions \union {abs(master_caa)})
         /\ master_caa' = step1(master_caa)
         /\ pc' = "Act"

Lock == /\ pc = "Lock"
        /\ master_caa' = -1 * abs(master_caa)
        /\ pc' = "Act"
        /\ UNCHANGED issued_sessions

Unlock == /\ pc = "Unlock"
          /\ master_caa' = abs(master_caa)
          /\ pc' = "Act"
          /\ UNCHANGED issued_sessions

Revoke == /\ pc = "Revoke"
          /\ master_caa' = step(master_caa, min(abs(master_caa), DELTA))
          /\ pc' = "Act"
          /\ UNCHANGED issued_sessions

Next == Act \/ Issue \/ Lock \/ Unlock \/ Revoke
           \/ (* Disjunct to prevent deadlock on termination *)
              (pc = "Done" /\ UNCHANGED vars)

Spec == Init /\ [][Next]_vars

Termination == <>(pc = "Done")

\* END TRANSLATION
=============================================================================
\* Modification History
\* Last modified Fri Mar 15 13:53:53 GMT 2019 by adrianduke
\* Created Sat Aug 12 16:10:55 BST 2017 by adrianduke

(*
------------------------------ MODULE TwoPhase_2 ------------------------------

(***************************************************************************)
(* This specification is discussed in "Two-Phase Commit", Lecture 6 of the *)
(* TLA+ Video Course.  It describes the Two-Phase Commit protocol, in      *)
(* which a transaction manager (TM) coordinates the resource managers      *)
(* (RMs) to implement the Transaction Commit specification of module       *)
(* TCommit.  In this specification, RMs spontaneously issue Prepared       *)
(* messages.  We ignore the Prepare messages that the TM can send to the   *)
(* RMs.                                                                    *)
(*                                                                         *)
(* For simplicity, we also eliminate Abort messages sent by an RM when it  *)
(* decides to abort.  Such a message would cause the TM to abort the       *)
(* transaction, an event represented here by the TM spontaneously deciding *)
(* to abort.                                                               *)
(***************************************************************************)
CONSTANT RM  \* The set of resource managers

VARIABLES
  rmState,       \* rmState[r] is the state of resource manager r.
  tmState,       \* The state of the transaction manager.
  tmPrepared,    \* The set of RMs from which the TM has received "Prepared"
                 \* messages.
  msgs           
    (***********************************************************************)
    (* In the protocol, processes communicate with one another by sending  *)
    (* messages.  For simplicity, we represent message passing with the    *)
    (* variable msgs whose value is the set of all messages that have been *)
    (* sent.  A message is sent by adding it to the set msgs.  An action   *)
    (* that, in an implementation, would be enabled by the receipt of a    *)
    (* certain message is here enabled by the presence of that message in  *)
    (* msgs.  For simplicity, messages are never removed from msgs.  This  *)
    (* allows a single message to be received by multiple receivers.       *)
    (* Receipt of the same message twice is therefore allowed; but in this *)
    (* particular protocol, that's not a problem.                          *)
    (***********************************************************************)

Messages ==
  (*************************************************************************)
  (* The set of all possible messages.  Messages of type "Prepared" are    *)
  (* sent from the RM indicated by the message's rm field to the TM.       *)
  (* Messages of type "Commit" and "Abort" are broadcast by the TM, to be  *)
  (* received by all RMs.  The set msgs contains just a single copy of     *)
  (* such a message.                                                       *)
  (*************************************************************************)
  [type : {"Prepared"}, rm : RM]  \cup  [type : {"Commit", "Abort"}]
   
TPTypeOK ==  
  (*************************************************************************)
  (* The type-correctness invariant                                        *)
  (*************************************************************************)
  /\ rmState \in [RM -> {"working", "prepared", "committed", "aborted"}]
  /\ tmState \in {"init", "done"}
  /\ tmPrepared \subseteq RM
  /\ msgs \subseteq Messages

TPInit ==   
  (*************************************************************************)
  (* The initial predicate.                                                *)
  (*************************************************************************)
  /\ rmState = [r \in RM |-> "working"]
  /\ tmState = "init"
  /\ tmPrepared   = {}
  /\ msgs = {}
-----------------------------------------------------------------------------
(***************************************************************************)
(* We now define the actions that may be performed by the processes, first *)
(* the TM's actions, then the RMs' actions.                                *)
(***************************************************************************)
TMRcvPrepared(r) ==
  (*************************************************************************)
  (* The TM receives a "Prepared" message from resource manager r.  We     *)
  (* could add the additional enabling condition r \notin tmPrepared,      *)
  (* which disables the action if the TM has already received this         *)
  (* message.  But there is no need, because in that case the action has   *)
  (* no effect; it leaves the state unchanged.                             *)
  (*************************************************************************)
  /\ tmState = "init"
  /\ [type |-> "Prepared", rm |-> r] \in msgs
  /\ tmPrepared' = tmPrepared \cup {r}
  /\ UNCHANGED <<rmState, tmState, msgs>>

TMCommit ==
  (*************************************************************************)
  (* The TM commits the transaction; enabled iff the TM is in its initial  *)
  (* state and every RM has sent a "Prepared" message.                     *)
  (*************************************************************************)
  /\ tmState = "init"
  /\ tmPrepared = RM
  /\ tmState' = "done"
  /\ msgs' = msgs \cup {[type |-> "Commit"]}
  /\ UNCHANGED <<rmState, tmPrepared>>

TMAbort ==
  (*************************************************************************)
  (* The TM spontaneously aborts the transaction.                          *)
  (*************************************************************************)
  /\ tmState = "init"
  /\ tmState' = "done"
  /\ msgs' = msgs \cup {[type |-> "Abort"]}
  /\ UNCHANGED <<rmState, tmPrepared>>

RMPrepare(r) == 
  (*************************************************************************)
  (* Resource manager r prepares.                                          *)
  (*************************************************************************)
  /\ rmState[r] = "working"
  /\ rmState' = [rmState EXCEPT ![r] = "prepared"]
  /\ msgs' = msgs \cup {[type |-> "Prepared", rm |-> r]}
  /\ UNCHANGED <<tmState, tmPrepared>>
  
RMChooseToAbort(r) ==
  (*************************************************************************)
  (* Resource manager r spontaneously decides to abort.  As noted above, r *)
  (* does not send any message in our simplified spec.                     *)
  (*************************************************************************)
  /\ rmState[r] = "working"
  /\ rmState' = [rmState EXCEPT ![r] = "aborted"]
  /\ UNCHANGED <<tmState, tmPrepared, msgs>>

RMRcvCommitMsg(r) ==
  (*************************************************************************)
  (* Resource manager r is told by the TM to commit.                       *)
  (*************************************************************************)
  /\ [type |-> "Commit"] \in msgs
  /\ rmState' = [rmState EXCEPT ![r] = "committed"]
  /\ UNCHANGED <<tmState, tmPrepared, msgs>>

RMRcvAbortMsg(r) ==
  (*************************************************************************)
  (* Resource manager r is told by the TM to abort.                        *)
  (*************************************************************************)
  /\ [type |-> "Abort"] \in msgs
  /\ rmState' = [rmState EXCEPT ![r] = "aborted"]
  /\ UNCHANGED <<tmState, tmPrepared, msgs>>

TPNext ==
  \/ TMCommit \/ TMAbort
  \/ \E r \in RM : 
       TMRcvPrepared(r) \/ RMPrepare(r) \/ RMChooseToAbort(r)
         \/ RMRcvCommitMsg(r) \/ RMRcvAbortMsg(r)
-----------------------------------------------------------------------------
(***************************************************************************)
(* The material below this point is not discussed in Video Lecture 6.  It  *)
(* will be explained in Video Lecture 8.                                   *)
(***************************************************************************)

TPSpec == TPInit /\ [][TPNext]_<<rmState, tmState, tmPrepared, msgs>>
  (*************************************************************************)
  (* The complete spec of the Two-Phase Commit protocol.                   *)
  (*************************************************************************)

THEOREM TPSpec => []TPTypeOK
  (*************************************************************************)
  (* This theorem asserts that the type-correctness predicate TPTypeOK is  *)
  (* an invariant of the specification.                                    *)
  (*************************************************************************)
-----------------------------------------------------------------------------
(***************************************************************************)
(* We now assert that the Two-Phase Commit protocol implements the         *)
(* Transaction Commit protocol of module TCommit.  The following statement *)
(* imports all the definitions from module TCommit into the current        *)
(* module.                                                                 *)
(***************************************************************************)
INSTANCE TCommit 

THEOREM TPSpec => TCSpec
  (*************************************************************************)
  (* This theorem asserts that the specification TPSpec of the Two-Phase   *)
  (* Commit protocol implements the specification TCSpec of the            *)
  (* Transaction Commit protocol.                                          *)
  (*************************************************************************)
(***************************************************************************)
(* The two theorems in this module have been checked with TLC for six      *)
(* RMs, a configuration with 50816 reachable states, in a little over a    *)
(* minute on a 1 GHz PC.                                                   *)
(***************************************************************************)

=============================================================================
\* Modification History
\* Last modified Wed Aug 16 11:48:34 BST 2017 by adrianduke
\* Created Tue Aug 15 16:47:08 BST 2017 by adrianduke

*)
