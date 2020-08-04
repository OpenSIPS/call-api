var callApiURL = "ws://164.90.215.149:5059/call-api";

var client = new $.JsonRpcClient({socketUrl: callApiURL, onmessage: CallOnMessage });

var callid = null;
var onHold = false;

function CallStart() {
	var caller = document.getElementById("caller").value;
	var callee = document.getElementById("callee").value;

	$('#callModal').on('hidden.bs.modal', function () {
		CallEnd();
	});
	toastr.info("Starting Call to " + caller);

	client.call("CallStart", { caller: caller, callee: callee},
		function(result) {
			toastr.info("Call initiated to caller");
			$('#callModal').find('.modal-title').text("Calling to " + caller);
			$('#caller-out').text(caller);
			$('#callee-out').text(callee);
		}, toastr.error);
}

function CallHold() {
	if (!callid) {
		toastr.error("Unknown call");
		return;
	}
	if (!onHold) {
		client.call("CallHold", { callid: callid },
			function(result) {
				toastr.info("Call on hold");
				$('#callModal').find('.modal-title').text("Call on hold")
				$('#callHold').addClass("active");
				$("#callHold").prop('aria-pressed', true)
				onHold = true;
			}, toastr.error);
	} else {
		client.call("CallUnhold", { callid: callid },
			function(result) {
				toastr.info("Call resumed");
				$('#callModal').find('.modal-title').text("Call resumed")
				$('#callHold').removeClass("active");
				$("#callHold").prop('aria-pressed', false)
				onHold = false;
			}, toastr.error);
	}
}

function CallEnd() {
	if (!callid) {
		toastr.error("Unknown call");
		return;
	}
	client.call("CallEnd", { callid: callid},
		function(result) {
			toastr.error("Call Ended");
			$('#callHold').addClass("invisible");
			$('#callEnd').addClass("invisible");
			$('#callid').text("");
			$('#callModal').modal('hide');
			callid = null;
			onHold = false;
		}, toastr.error);
}

function CallOnMessage(message) {
	var obj = JSON.parse(message.data);
	switch(obj.params.event) {
		case "Error":
			toastr.error(obj.params.data);
			$('#callModal').modal('hide');
			break;
		case "CallerAnswered":
			toastr.info("Caller Answered");
			$('#callModal').find('.modal-title').text(obj.params.data.caller + " answered");
			break;
		case "TransferStart":
			toastr.info("Calling Callee");
			$('#callModal').find('.modal-title').text("Calling to " + obj.params.data.callee);
			callid = obj.params.data.callid;
			$('#callid').text(callid);
			break;
		case "TransferPending":
			toastr.info("Callee status: " + obj.params.data.extra);
			break;
		case "Transferring":
			toastr.info("Transferring");
			break;
		case "CalleeAnswered":
			toastr.success("Callee Answered");
			$('#callModal').find('.modal-title').text("Ongoing call");
			$('#callHold').removeClass("invisible");
			$('#callEnd').removeClass("invisible");
			break;
		case "CallHolding":
		case "CallUnholding":
		case "Ended":
			break;
		default:
			toastr.warning("Unhandled message: " + obj.params.data);
			console.log(obj.params);
			break;
	}
}
